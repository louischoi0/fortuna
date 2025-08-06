package test

import (
	"bytes"
	"fmt"
	"fortuna/core/model"
	"testing"
)

func TestTransaction(t *testing.T) {
	t.Run("test transaction hash", func(t *testing.T) {

		transaction := model.NewRawTransaction("space_id", "from")
		transaction.Timestamp = 1754354451836053510

		transaction.AddOperation(model.NewOperation("x0", "write_var", []interface{}{1}))
		transaction.UpdateRawCode()

		hash := transaction.Hash()
		fmt.Println(hash)
	})
}

func TestRawTransaction(t *testing.T) {
	t.Run("encode/decode and verify hash", func(t *testing.T) {
		// Given: Create a raw transaction
		rtx := model.NewRawTransaction("space_id___________________________________________1234560000000", "from_address________________________________________567800000000")
		rtx.Timestamp = 1754354451836053510
		rtx.AddOperation(model.NewOperation("x0", "write_var", []interface{}{2}))

		// When: encode the transaction
		encoded, err := rtx.Encode()
		if err != nil {
			t.Fatalf("failed to encode transaction: %v", err)
		}

		// And: decode the transaction
		decoded, err := model.DecodeRawTransaction(encoded)
		if err != nil {
			t.Fatalf("failed to decode transaction: %v", err)
		}

		// Then: verify original and decoded transaction match
		if rtx.Timestamp != decoded.Timestamp {
			t.Errorf("timestamp mismatch: got %d, want %d", decoded.Timestamp, rtx.Timestamp)
		}
		if rtx.SpaceID != decoded.SpaceID {
			t.Errorf("space_id mismatch: got %s, want %s", decoded.SpaceID, rtx.SpaceID)
		}
		if rtx.From != decoded.From {
			t.Errorf("from mismatch: got %s, want %s", decoded.From, rtx.From)
		}
		if !bytes.Equal(rtx.RawCode, decoded.RawCode) {
			t.Errorf("raw_code mismatch: got %v, want %v", decoded.RawCode, rtx.RawCode)
		}
		if len(decoded.Operations) != 1 {
			t.Errorf("expected 1 operation, got %d", len(decoded.Operations))
		}
		if decoded.Operations[0].OpCode != "x0" {
			t.Errorf("operation object mismatch: got %s", decoded.Operations[0].OpCode)
		}

		// Finally: hash match
		if rtx.Hash() != decoded.Hash() {
			t.Errorf("hash mismatch: encoded=%s, decoded=%s", rtx.Hash(), decoded.Hash())
		}

		fmt.Println(rtx.Hash())
		fmt.Println(decoded.Hash())
		fmt.Println(string(rtx.RawCode))
		fmt.Println(string(decoded.RawCode))
	})
}
