package test

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/structure"
	"testing"
)

func NewTestChain() *model.Chain {
	chain := model.NewChain("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	return chain
}

func NewTestBlock1(h int64) *model.Block {
	timestamp := 1754354451836053510
	ebuf := fmt.Sprintf(`{"timestamp": %v, "space_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","publisher": "0000000000000000000000000000000000000000000000000000000000000000","payload":{"a":3},"spec":{"interface_id":"FIC-00-00001","version":"base::v0.0.0","params":{"slot_count":64}},"topic":"","subtopic":"","seperator":"","tag":""}`, timestamp)
	om, _ := structure.ParseOrderedMap(ebuf)
	event, _ := model.NewEventFromOrderedMap(om)
	exec := model.NewEventResultFromEvent(event, "success")
	block := model.NewBlock("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", h, nil)

	block.AppendEventExecution(exec)
	block.Timestamp = int64(timestamp)

	block.UpdateTransactionRoot()
	block.UpdateExecutionRoot()

	return block
}

func TestChain(t *testing.T) {
	chain := NewTestChain()

	height := int64(1)
	block := NewTestBlock1(height)
	err := chain.CommitBlock(block)
	
	if err != nil {
		t.Fatalf("failed to commit block: %v", err)
	}

	block, err = chain.ReadBlock(height)
	if err != nil {
		t.Fatalf("failed to get block: %v", err)
	}

	if block == nil {
		t.Fatalf("block is nil")
	}

	err = chain.Storage.Close()

	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	loaded_chain := NewTestChain()

	loaded_chain.LoadChainData()
	loaded_chain.Storage.DeleteAll()

	if loaded_chain.LastHeight != block.Height {
		t.Fatalf("expected chain lastheight %v, but %v", 1, loaded_chain.LastHeight)
	}

	block, err = loaded_chain.ReadBlock(1)
	if err != nil {
		t.Fatalf("failed to get block: %v", err)
	}

	if block == nil {
		t.Fatalf("block is nil")
	}

	if block.Height != 1 {
		t.Fatalf("expected block height %v, but %v", 1, block.Height)
	}

	block2 := NewTestBlock1(int64(2))
	err = loaded_chain.CommitBlock(block2)
	
	if err != nil {
		t.Fatalf("failed to commit block: %v", err)
	}

	loaded_block2, err := loaded_chain.ReadBlock(2)

	if block2.Height != loaded_block2.Height {
		t.Fatalf("height err")
	}

	if block2.Hash() != loaded_block2.Hash() {
		t.Fatalf("hash err")
	}

	if err := loaded_chain.Reset(); err != nil {
		t.Fatalf("failed to reset storage: %s", err.Error())
	}

}
