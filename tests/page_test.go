package test

import (
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/structure"
	"testing"
	"fmt"
)

func TestPageEncodeDecode(t *testing.T) {
	// Create a test page
	spaceID := "0000000000000000000000000000000000000000000000000000000000000000"
	page := model.NewPage(spaceID, 1, nil)

	// Add some test executions
	params := structure.NewOrderedMap()
	params.Set("test_param", "test_value")

	payload := structure.NewOrderedMap()
	payload.Set("test_payload", "test_value")

	spec := model.NewEventSpec("interface0", "kernel0000", params)

	event, err := model.NewEventRequest(
		"0000000000000000000000000000000000000000000000000000000000000000",
		spaceID,
		payload,
		spec,
		"test_topic",
		"test_subtopic",
		"test_tag",
	)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	eventResult := model.NewEventResultFromEvent(event, "test_result")
	page.AppendEventExecution(eventResult)

	// Add some test transactions
	v := model.NewStateVector(make([]int64, 8, 8))
	fmt.Println(v.Data)

	transaction := model.NewRawTransaction("space_id", "from")
	transaction.SpaceID = "0000000000000000000000000000000000000000000000000000000000000000"
	transaction.From = "0000000000000000000000000000000000000000000000000000000000000000"
	transaction.Timestamp = 1754354451836053510
	abi := vm.ABI{}
	op := abi.WriteVar("k", v.Var())

	transaction.AddOperation(op)
	transaction.UpdateRawCode()

	page.AppendTransaction(transaction)

	// Update roots
	page.UpdateExecutionRoot()
	page.UpdateTransactionRoot()

	// Test Encode
	encoded, err := page.Encode()
	if err != nil {
		t.Fatalf("Failed to encode page: %v", err)
	}

	// Test Decode
	decoded, err := model.DecodePage(encoded)
	if err != nil {
		t.Fatalf("Failed to decode page: %v", err)
	}


	if len(decoded.Transactions) != 1 {
		t.Fatalf("transactions expected have len: %d", 1)
	}

	a := decoded.Transactions[0].Operations[0].Args[1]
	aa, ok := a.(*model.Var)
	if !ok {
		t.Fatalf("expected to have var type")
	}
	sv := aa.Vec()
	fmt.Println(sv.Data)

	// Verify decoded values
	t.Run("Basic Fields", func(t *testing.T) {
		if decoded.N != page.N {
			t.Errorf("Page number mismatch: got %v, want %v", decoded.N, page.N)
		}
		if decoded.Timestamp != page.Timestamp {
			t.Errorf("Timestamp mismatch: got %v, want %v", decoded.Timestamp, page.Timestamp)
		}
		if decoded.PrevPageHash != page.PrevPageHash {
			t.Errorf("PrevPageHash mismatch: got %v, want %v", decoded.PrevPageHash, page.PrevPageHash)
		}
	})

	/**
	t.Run("Root Hashes", func(t *testing.T) {
		if decoded.TransactionRootHash != page.TransactionRootHash {
			t.Errorf("TransactionRootHash mismatch: got %v, want %v", decoded.TransactionRootHash, page.TransactionRootHash)
		}
		if decoded.ExecutionRootHash != page.ExecutionRootHash {
			t.Errorf("ExecutionRootHash mismatch: got %v, want %v", decoded.ExecutionRootHash, page.ExecutionRootHash)
		}
	})
	**/

	t.Run("Executions", func(t *testing.T) {
		if len(decoded.Executions) != len(page.Executions) {
			t.Errorf("Executions count mismatch: got %v, want %v", len(decoded.Executions), len(page.Executions))
		}
		if len(decoded.Executions) > 0 {
			if decoded.Executions[0].Result != page.Executions[0].Result {
				t.Errorf("Execution result mismatch: got %v, want %v", decoded.Executions[0].Result, page.Executions[0].Result)
			}
		}
	})

	t.Run("Transactions", func(t *testing.T) {
		if len(decoded.Transactions) != len(page.Transactions) {
			t.Errorf("Transactions count mismatch: got %v, want %v", len(decoded.Transactions), len(page.Transactions))
		}
		if len(decoded.Transactions) > 0 {
			if decoded.Transactions[0].SpaceID != page.Transactions[0].SpaceID {
				t.Errorf("Transaction SpaceID mismatch: got %v, want %v", decoded.Transactions[0].SpaceID, page.Transactions[0].SpaceID)
			}
		}
	})

	/**
	t.Run("Hash Verification", func(t *testing.T) {
		if decoded.Hash() != page.Hash() {
			t.Errorf("Hash mismatch: got %v, want %v", decoded.Hash(), page.Hash())
		}
	})
	**/
}

/**
func TestPageIndex_EncodeDecode(t *testing.T) {
	// Create test data
	spaceID := "space_id_000000000000000000000000000000000000000000000000000000000000"
	pageNum := uint64(1)
	blockNum := uint64(1)
	offset := uint64(0)
	size := uint64(100)

	// Create PageIndex
	pageIndex := model.NewPageIndex(spaceID, pageNum, blockNum, offset, size)

	// Test Encode
	encoded, err := pageIndex.Encode()
	if err != nil {
		t.Fatalf("Failed to encode PageIndex: %v", err)
	}

	// Test Decode
	decoded, err := model.DecodePageIndex(encoded)
	if err != nil {
		t.Fatalf("Failed to decode PageIndex: %v", err)
	}

	// Verify decoded values
	t.Run("Basic Fields", func(t *testing.T) {
		if decoded.SpaceID != spaceID {
			t.Errorf("SpaceID mismatch: got %v, want %v", decoded.SpaceID, spaceID)
		}
		if decoded.PageNum != pageNum {
			t.Errorf("PageNum mismatch: got %v, want %v", decoded.PageNum, pageNum)
		}
		if decoded.BlockNum != blockNum {
			t.Errorf("BlockNum mismatch: got %v, want %v", decoded.BlockNum, blockNum)
		}
		if decoded.Offset != offset {
			t.Errorf("Offset mismatch: got %v, want %v", decoded.Offset, offset)
		}
		if decoded.Size != size {
			t.Errorf("Size mismatch: got %v, want %v", decoded.Size, size)
		}
	})
}
**/
