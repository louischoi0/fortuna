package test

import (
	"fortuna/core/model"
	"testing"
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
	})
}
