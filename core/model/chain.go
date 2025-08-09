package model

import (
	"encoding/json"
	"fmt"
	"fortuna/core/storage"
	"fortuna/rock"
	"fortuna/util"
	"os"
	"path/filepath"
	"time"

	"github.com/linxGnu/grocksdb"
)

const BLOCK_FILE_SIZE = 1024 * 1024 * 256

type blockIndex struct {
	Height int64
	FileNo int64
	Offset int64
	Size   int64
}

func (b *blockIndex) Encode() ([]byte, error) {
	// TODO
	return json.Marshal(b)
}

func (b *blockIndex) Decode(data []byte) error {
	//TODO
	return json.Unmarshal(data, b)
}

func NewBlockIndex(height int64, fileNo int64, offset int64, size int64) *blockIndex {
	return &blockIndex{
		Height: height,
		FileNo: fileNo,
		Offset: offset,
		Size:   size,
	}
}

type Chain struct {
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

func NewChain(metaDB *grocksdb.DB, spaceID string) *Chain {
	dir := filepath.Join(storage.DataRootDir(), spaceID)
	os.MkdirAll(dir, 0755)
	storage := storage.NewFileStorage(dir)

	return &Chain{
		meta:    metaDB,
		Storage: storage,
	}
}

func GetFileName(fileNo int64) string {
	return fmt.Sprintf("blk%06d", fileNo)
}

func (c *Chain) NewNextBlockFile() error {
	c.FileNo++
	fileName := GetFileName(c.FileNo)
	c.CurrentFile = storage.NewFile(c.Storage, fileName)
	return nil
}

func (c *Chain) IsFileSizeExceed(f *storage.File) bool {
	return f.GetSize() >= BLOCK_FILE_SIZE
}

func (c *Chain) CommitBlock(block *Block) error {

	if c.LastHeight+1 != block.Height {
		return fmt.Errorf("height %v is exepcted not %v", c.LastHeight+1, block.Height)
	}

	lastBlockHash := c.LastBlock.Hash()

	if block.Height != 1 && lastBlockHash != block.PreviousBlockHash {
		return fmt.Errorf("previous block hash mismatch expected %s not %s", lastBlockHash, block.PreviousBlockHash)
	}

	if c.IsFileSizeExceed(c.CurrentFile) {
		c.CurrentFile.Close()
		c.NewNextBlockFile()
	}

	block_buffer, err := block.Encode()
	if err != nil {
		return err
	}

	blockIndex := NewBlockIndex(block.Height, c.FileNo, c.CurrentFile.GetSize(), int64(len(block_buffer)))
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

func (c *Chain) GetBlockIndex(height int64) *blockIndex {
	key := fmt.Sprintf("blkidx:%d", height)
	value, err := rock.GetValue(c.meta, key)
	if err != nil {
		return nil
	}

	return value.(*blockIndex)
}

func (c *Chain) WriteBlockIndex(blockIndex *blockIndex) error {
	key := fmt.Sprintf("blkidx:%d", blockIndex.Height)
	blockIndexBytes, err := blockIndex.Encode()
	if err != nil {
		return err
	}
	rock.SetValue(c.meta, key, blockIndexBytes)

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
