package model

import (
	"bytes"
	"fmt"
	"fortuna/core/storage"
	"fortuna/rock"
	"fortuna/util"
	"log"
	"path/filepath"
	"sync"
	"time"

	C "fortuna/core/config"

	"github.com/linxGnu/grocksdb"
)

type BlockIndex struct {
	SpaceID string
	Height  int64
	FileNo  int64
	Offset  int64
	Size    int64
}

func (b *BlockIndex) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(b.SpaceID)

	buffer.Write(util.EncodeUint64(uint64(b.Height)))
	buffer.Write(util.EncodeUint64(uint64(b.FileNo)))
	buffer.Write(util.EncodeUint64(uint64(b.Offset)))
	buffer.Write(util.EncodeUint64(uint64(b.Size)))

	return buffer.Bytes(), nil
}

func DecodeBlockIndex(data []byte) (*BlockIndex, error) {
	b := &BlockIndex{}

	var (
		off = 0
		n   = len(data)
	)

	need := func(k int) error {
		if off+k > n {
			return fmt.Errorf("buffer underflow: need %d bytes, value: %s", k, string(data))
		}
		return nil
	}
	readFixedString := func(k int) (string, error) {
		if err := need(k); err != nil {
			return "", err
		}
		s := string(data[off : off+k])
		off += k
		return s, nil
	}
	readU64LE := func() (uint64, error) {
		if err := need(8); err != nil {
			return 0, err
		}
		v, err := util.DecodeUint64(data[off : off+8])
		if err != nil {
			return 0, err
		}
		off += 8
		return v, nil
	}

	spaceID, err := readFixedString(64)
	if err != nil {
		return nil, err
	}
	b.SpaceID = spaceID

	height, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.Height = int64(height)

	fileNo, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.FileNo = int64(fileNo)

	offset, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.Offset = int64(offset)

	size, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.Size = int64(size)

	return b, nil
}

func NewBlockIndex(spaceID string, height int64, fileNo int64, offset int64, size int64) *BlockIndex {
	if len(spaceID) != 64 {
		log.Fatalf("spaceID has invalid length %s", spaceID)
	}

	return &BlockIndex{
		SpaceID: spaceID,
		Height:  height,
		FileNo:  fileNo,
		Offset:  offset,
		Size:    size,
	}
}

type Chain struct {
	mu        sync.Mutex
	SpaceID   string
	LastBlock *Block
	Blocks    []*Block

	Storage *storage.FileStorage
	meta    *grocksdb.DB

	FileNo      int64
	CurrentFile *storage.File

	LastHeight     int64
	LastAppendedAt time.Time
}

func NewChain(spaceID string) *Chain {
	dir := filepath.Join(storage.DataRootDir(), fmt.Sprintf("chain.%s", spaceID))
	storage := storage.NewFileStorage(dir)
	metaDB, err := rock.GetDBInstance(fmt.Sprintf("chain_additional.%s", spaceID))
	if err != nil {
		log.Fatalf(err.Error())
	}

	return &Chain{
		SpaceID: spaceID,
		meta:    metaDB,
		Storage: storage,
	}
}

func (c *Chain) LoadChainData() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	height := c.ReadLastHeight()
	c.LastHeight = height

	if len(c.Blocks) != 0 {
		return fmt.Errorf("chain loaded again")
	}

	for idx := range height {
		blk, err := c.ReadBlock(idx)
		if err != nil {
			return err
		}
		c.Blocks = append(c.Blocks, blk)
	}

	return nil
}

func GetFileName(fileNo int64) string {
	return fmt.Sprintf(C.BLOCK_FILE_NAME_FORMAT, fileNo)
}

func (c *Chain) NewNextBlockFile() error {
	c.FileNo++
	fileName := GetFileName(c.FileNo)
	c.CurrentFile = storage.NewFile(c.Storage, fileName)

	err := c.SetCurrentFile(c.CurrentFile)
	if err != nil {
		return err
	}

	return nil
}

func (c *Chain) IsFileSizeExceed(f *storage.File) bool {
	return f.GetSize() >= C.BLOCK_FILE_SIZE
}

func (c *Chain) CommitBlock(block *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if h := c.ReadLastHeight(); h != c.LastHeight {
		return fmt.Errorf("chain height mismatch expected %v, but %v", c.LastHeight, h)
	}

	if c.LastHeight+1 != block.Height {
		return fmt.Errorf("height %v is exepcted not %v", c.LastHeight+1, block.Height)
	}

	if block.Height != 1 && c.LastBlock != nil && c.LastBlock.Hash() != block.PreviousBlockHash {
		return fmt.Errorf("previous block hash mismatch expected %s not %s", c.LastBlock.Hash(), block.PreviousBlockHash)
	}
	if block.Height == 1 && c.LastBlock != nil {
		return fmt.Errorf("first block must be genesis block")
	}
	if block.Height != 1 && c.LastBlock == nil {
		return fmt.Errorf("last block is nil")
	}

	if c.CurrentFile == nil || c.IsFileSizeExceed(c.CurrentFile) {
		c.NewNextBlockFile()
	}

	block_buffer, err := block.Encode()
	if err != nil {
		return err
	}

	blockIndex := NewBlockIndex(c.SpaceID, block.Height, c.FileNo, c.CurrentFile.GetSize(), int64(len(block_buffer)))
	err = c.WriteBlockIndex(blockIndex)
	if err != nil {
		return err
	}

	err = c.CurrentFile.AppendFileBytes(block_buffer)
	if err != nil {
		return err
	}

	err = c.SetHeight(block.Height)
	if err != nil {
		return err
	}

	c.LastBlock = block
	c.LastAppendedAt = time.Now()

	return nil
}

func (c *Chain) GetBlock(height int64) (*Block, error) {

	if int(height) > len(c.Blocks)+1 {
		return nil, fmt.Errorf("block not found for height: %v", height)
	}

	blk := c.Blocks[height-1]

	if blk.Height != height {
		log.Fatalf("block invalid error block expected height: %v, but %v", blk.Height, height)
	}

	return blk, nil
}

func (c *Chain) ReadBlock(height int64) (*Block, error) {
	blockIndex, err := c.ReadBlockIndex(height)
	if err != nil {
		return nil, fmt.Errorf("failed to get BlockIndex: %v", err.Error())
	}

	fileNo := blockIndex.FileNo
	fileName := GetFileName(fileNo)
	file := c.Storage.GetOrOpenFile(fileName)

	fmt.Printf("offset: %v size: %v", blockIndex.Offset, blockIndex.Size)

	blockBytes, err := file.OffsetRead(blockIndex.Offset, blockIndex.Size)
	if err != nil {
		return nil, err
	}

	block, err := DecodeBlock(blockBytes)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (c *Chain) ReadBlockIndex(height int64) (*BlockIndex, error) {
	key := fmt.Sprintf(C.BLOCK_INDEX_KEY_FORMAT, height)
	value, err := rock.GetValue(c.meta, key)
	if err != nil {
		return nil, err
	}

	blockIndex, err := DecodeBlockIndex(value)
	if err != nil {
		return nil, err
	}

	return blockIndex, nil
}

func (c *Chain) WriteBlockIndex(blockIndex *BlockIndex) error {
	key := fmt.Sprintf(C.BLOCK_INDEX_KEY_FORMAT, blockIndex.Height)
	blockIndexBytes, err := blockIndex.Encode()
	if err != nil {
		return err
	}

	rock.SetValue(c.meta, key, blockIndexBytes)
	log.Printf("write block index at height %v - size:%v offset: %v buffer_size: %v", blockIndex.Height, blockIndex.Size, blockIndex.Offset, len(blockIndexBytes))
	return nil
}

func (c *Chain) SetCurrentFile(file *storage.File) error {
	err := rock.SetValue(c.meta, "current", []byte(file.Name))
	if err != nil {
		return err
	}

	c.CurrentFile = file
	return nil
}

func (c *Chain) SetHeight(height int64) error {
	err := rock.SetValue(c.meta, "height", []byte(util.EncodeUint64(uint64(height))))
	if err != nil {
		return err
	}
	c.LastHeight = height
	return nil
}

func (c *Chain) ReadLastHeight() int64 {
	height, err := rock.GetValue(c.meta, "height")
	if err != nil {
		return 0
	}
	h, err := util.DecodeUint64(height)
	if err != nil {
		return 0
	}
	return int64(h)
}

func (c *Chain) GetHeight() int64 {
	return c.LastHeight
}

func (c *Chain) Next() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Printf("next height: %v", c.LastHeight+1)

	return c.LastHeight + 1
}

func (c *Chain) SetMetaDB(meta *grocksdb.DB) {
	c.meta = meta
}
