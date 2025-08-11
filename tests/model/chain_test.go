package test

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/structure"
	"testing"
)

func NewTestChain() *model.Chain {
	chain := model.NewChain("test")
	return chain
}

func NewTestBlock1() *model.Block {
	timestamp := 1754354451836053510
	ebuf := fmt.Sprintf(`{"timestamp": %v, "space_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","publisher": "0000000000000000000000000000000000000000000000000000000000000000","payload":{"a":3},"spec":{"interface_id":"FIC-00-00001","version":"base::v0.0.0","params":{"slot_count":64}},"topic":"","subtopic":"","seperator":"","tag":""}`, timestamp)
	om, _ := structure.ParseOrderedMap(ebuf)
	event, _ := model.NewEventFromOrderedMap(om)
	exec := model.NewEventResultFromEvent(event, "success")
	block := model.NewBlock("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1, nil)

	block.AppendEventExecution(exec)
	block.Timestamp = int64(timestamp)

	block.UpdateTransactionRoot()
	block.UpdateExecutionRoot()

	return block
}

func TestChain(t *testing.T) {
	chain := NewTestChain()
	block := NewTestBlock1()
	chain.CommitBlock(block)

	block, err := chain.GetBlock(1)
	if err != nil {
		t.Fatalf("failed to get block: %v", err)
	}

	if block == nil {
		t.Fatalf("block is nil")
	}
}
