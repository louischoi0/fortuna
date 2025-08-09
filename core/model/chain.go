package model

import (
	"time"
	"fmt"
	"fortuna/core/storage"
	"github.com/linxGnu/grocksdb"
)

type Chain struct {
	SpaceID 	string
	LastBlock	*Block
	Blocks  	[]*Block

	Storage		*storage.FileStorage
	meta		*grocksdb.DB

	FileNo		int64
	CurrentFile	*storage.File

	LastHeight     	int64
	LastAppendedAt 	time.Time
	LastVerifiedAt 	time.Time
}

func (c *Chain) NewNextBlockFile() error {
	c.FileNo++
	fileName := fmt.Sprintf("blk%06d", c.FileNo)
	c.CurrentFile = storage.NewFile(c.Storage, fileName)
	return nil
}

func (c *Chain) CheckCurrentFileFull() bool {
	return false
}

func (c *Chain) CommitBlock(block *Block) error {

	if c.LastHeight + 1 != block.Height {
		return fmt.Errorf("height %v is exepcted not %v", c.LastHeight+1, block.Height)
	}
	
	lastBlockHash := c.LastBlock.Hash()

	if block.Height != 1 && lastBlockHash != block.PreviousBlockHash {
		return fmt.Errorf("previous block hash mismatch expected %s not %s", lastBlockHash, block.PreviousBlockHash )
	}
	if c.CheckCurrentFileFull() {
		c.NewNextBlockFile()
	}

	block_buffer, err := block.Encode()
	if err != nil {
		return err
	}

	err = c.CurrentFile.AppendFileBytes(block_buffer)
	if err != nil {
		return err
	}

	return nil
}
