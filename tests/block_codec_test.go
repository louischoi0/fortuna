package test

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/structure"
	"testing"
)

func TestBlockCodec(t *testing.T) {
	t.Run("encode/decode block", func(t *testing.T) {
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

		exec := model.NewEventResultFromEvent(event, "success")
		block := model.NewBlock("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1, nil)
		block.AppendEventExecution(exec)
		block.Timestamp = int64(timestamp)
		block.UpdateTransactionRoot()
		block.UpdateExecutionRoot()

		blockBuffer, err := block.Encode()
		if err != nil {
			t.Fatalf("failed to encode block: %v", err)
		}

		decoded_block, err := model.DecodeBlock(blockBuffer)
		if err != nil {
			t.Fatalf("failed to decode block: %v", err)
		}

		if decoded_block.Timestamp != block.Timestamp {
			t.Fatalf("block timestamp mismatch: %d != %d", decoded_block.Timestamp, block.Timestamp)
		}

		if decoded_block.Height != block.Height {
			t.Fatalf("block height mismatch: %d != %d", decoded_block.Height, block.Height)
		}

		if decoded_block.TransactionRootHash != block.TransactionRootHash {
			t.Fatalf("block transaction root hash mismatch: %s != %s", decoded_block.TransactionRootHash, block.TransactionRootHash)
		}

		if decoded_block.ExecutionRootHash != block.ExecutionRootHash {
			t.Fatalf("block execution root hash mismatch: %s != %s", decoded_block.ExecutionRootHash, block.ExecutionRootHash)
		}

		if decoded_block.Hash() != block.Hash() {
			t.Fatalf("block hash mismatch: %s != %s", decoded_block.Hash(), block.Hash())
		}
	})
}
