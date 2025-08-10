package model

import (
	"encoding/json"
	"fmt"
	"fortuna/core/storage"
	"fortuna/rock"
	"fortuna/util"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/linxGnu/grocksdb"
)

const BLOCK_FILE_SIZE = 1024 * 1024 * 256
const BLOCK_FILE_NAME_FORMAT = "blk%06d"
const BLOCK_INDEX_KEY_FORMAT = "blkidx:%d"

type blockIndex struct {
	SpaceID string
	Height  int64
	FileNo  int64
	Offset  int64
	Size    int64
}

func (b *blockIndex) Encode() ([]byte, error) {
	// TODO
	return json.Marshal(b)
}

func (b *blockIndex) Decode(data []byte) error {
	//TODO
	return json.Unmarshal(data, b)
}

func NewBlockIndex(spaceID string, height int64, fileNo int64, offset int64, size int64) *blockIndex {
	return &blockIndex{
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
	os.MkdirAll(dir, 0755)
	storage := storage.NewFileStorage(dir)
	metaDB, err := rock.GetDBInstance(fmt.Sprintf("chain_additional.%s", spaceID))
	if err != nil {
		log.Fatalf(err.Error())
	}

	return &Chain{
		meta:    metaDB,
		Storage: storage,
	}
}

func (c *Chain) LoadChainData() error {
	height, err := c.GetHeight()
	if err != nil {
		return err
	}

	c.LastHeight = height

	return nil
}

func GetFileName(fileNo int64) string {
	return fmt.Sprintf(BLOCK_FILE_NAME_FORMAT, fileNo)
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
	return f.GetSize() >= BLOCK_FILE_SIZE
}

func (c *Chain) CommitBlock(block *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.LastHeight+1 != block.Height {
		return fmt.Errorf("height %v is exepcted not %v", c.LastHeight+1, block.Height)
	}

	lastBlockHash := c.LastBlock.Hash()

	if block.Height != 1 && lastBlockHash != block.PreviousBlockHash {
		return fmt.Errorf("previous block hash mismatch expected %s not %s", lastBlockHash, block.PreviousBlockHash)
	}

	if c.IsFileSizeExceed(c.CurrentFile) || c.CurrentFile == nil {
		if c.CurrentFile != nil {
			c.CurrentFile.Close()
		}
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
	blockIndex := c.GetBlockIndex(height)
	if blockIndex == nil {
		return nil, fmt.Errorf("block index not found")
	}

	fileName := GetFileName(blockIndex.FileNo)
	file := c.Storage.GetFile(fileName)
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

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

func (c *Chain) GetBlockIndex(height int64) *blockIndex {
	key := fmt.Sprintf(BLOCK_INDEX_KEY_FORMAT, height)
	value, err := rock.GetValue(c.meta, key)
	if err != nil {
		return nil
	}

	return value.(*blockIndex)
}

func (c *Chain) WriteBlockIndex(blockIndex *blockIndex) error {
	key := fmt.Sprintf(BLOCK_INDEX_KEY_FORMAT, blockIndex.Height)
	blockIndexBytes, err := blockIndex.Encode()
	if err != nil {
		return err
	}
	rock.SetValue(c.meta, key, blockIndexBytes)

	return nil
}

func (c *Chain) SetCurrentFile(file *storage.File) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := rock.SetValue(c.meta, "current", []byte(file.Name))
	if err != nil {
		return err
	}

	c.CurrentFile = file
	return nil
}

func (c *Chain) SetHeight(height int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := rock.SetValue(c.meta, "height", []byte(util.EncodeUint64(uint64(height))))
	if err != nil {
		return err
	}
	c.LastHeight = height
	return nil
}

func (c *Chain) GetHeight() (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	height, err := rock.GetValue(c.meta, "height")
	if err != nil {
		return 0, err
	}
	return int64(height.(uint64)), nil
}
