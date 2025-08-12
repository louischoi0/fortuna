package test

import (
	"fortuna/core/model"
	"fortuna/rock"
	"testing"
	"fmt"
)

func NewTestBlockIndex() *model.BlockIndex {
	return &model.BlockIndex{
		SpaceID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Height:  1,
		FileNo:  1,
		Offset:  1,
		Size:    1,
	}
}

func TestBlockIndexCodec(t *testing.T) {
	t.Run("encode/decode block index", func(t *testing.T) {
		testDB, err := rock.GetDBInstance("test")
		if err != nil {
			t.Fatalf("failed to get test db: %v", err)
		}
		defer rock.CloseDB(testDB)

		blockIndex := NewTestBlockIndex()
		encoded, err := blockIndex.Encode()
		if err != nil {
			t.Fatalf("failed to encode block index: %v", err)
		}
		decoded, err := model.DecodeBlockIndex(encoded)
		if err != nil {
			t.Fatalf("failed to decode block index: %v", err)
		}
		if decoded.SpaceID != blockIndex.SpaceID {
			t.Fatalf("space id mismatch: %v != %v", decoded.SpaceID, blockIndex.SpaceID)
		}
		if decoded.Height != blockIndex.Height {
			t.Fatalf("height mismatch: %v != %v", decoded.Height, blockIndex.Height)
		}

		c := &model.Chain{}
		c.SetMetaDB(testDB)

		err = c.WriteBlockIndex(blockIndex)
		if err != nil {
			t.Fatalf("failed to write block index: %v", err)
		}

		blockIndexRead, err := c.ReadBlockIndex(blockIndex.Height)

		fmt.Println("biread: ", blockIndexRead.Height)
		if err != nil {
			t.Fatalf("failed to get block index: %v", err)
		}

		fmt.Println("SpaceID: ", blockIndexRead.SpaceID[:64])

		if blockIndexRead.SpaceID != blockIndex.SpaceID {
			fmt.Printf("space id mismatch: %v != %v", blockIndexRead.SpaceID, blockIndex.SpaceID)
		}

		if blockIndexRead.Height != blockIndex.Height {
			fmt.Printf("height mismatch: %v != %v", blockIndexRead.Height, blockIndex.Height)
		}

		if blockIndexRead.Size != blockIndex.Size {
			fmt.Printf("height mismatch: %v != %v", blockIndexRead.Size, blockIndex.Size)
		}
	})
}
