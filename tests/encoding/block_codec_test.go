package test

import (
	"testing"
	"fmt"
	"fortuna/structure"
	"fortuna/core/model"
)


func TestEventExecutionResult(t *testing.T) {
	t.Run("encode/decode execution result with event", func(t *testing.T) {
		// GIVEN: an event with basic info
		timestamp := 1754354451836053510
		ebuf := fmt.Sprintf(`{"timestamp": %v, "space_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","publisher": "0000000000000000000000000000000000000000000000000000000000000000","payload":{"a":3},"spec":{"interface_id":"FIC-00-00001","version":"base::v0.0.0","params":{"slot_count":64}},"topic":"","subtopic":"","seperator":"","tag":""}`, timestamp)
		om, err := structure.ParseOrderedMap(ebuf)
		if err != nil {
			t.Fatalf("failed to parse ordered map: %v", err)
		}

		event, err := model.NewEventFromOrderedMap(om)
		if err != nil {
			t.Fatalf("failed to parse ordered map: %v", err)
		}

		exec := model.NewEventExecutionResultFromEvent(event, "success")
		block := model.NewBlock("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1, nil)
		block.AppendEventExecution(exec)

		blockBuffer, err := block.Encode()
		if err != nil {
			t.Fatalf("failed to encode block: %v", err)
		}
		
		fmt.Printf("origin block hash: %s", block.Hash())
		fmt.Printf("origin block: %v", block)

		decoded_block, err := model.DecodeBlock(blockBuffer)
		if err != nil {
			t.Fatalf("failed to decode block: %v", err)
		}

		fmt.Println(block)
		fmt.Println(decoded_block)
	})
}


