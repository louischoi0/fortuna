package test

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/structure"
	"fortuna/util"
	"testing"
)

func TestEncodeDecodeEvent(t *testing.T) {
	timestamp := util.Now()

	ebuf := fmt.Sprintf(`{"timestamp": %v, "space_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","publisher": "0000000000000000000000000000000000000000000000000000000000000000","payload":{"a":3},"spec":{"interface_id":"FIC-00-00001","version":"base::v0.0.0","params":{"slot_count":64}},"topic":"","subtopic":"","seperator":"","tag":""}`, timestamp)
	om, err := structure.ParseOrderedMap(ebuf)
	if err != nil {
		t.Fatalf("failed to parse ordered map: %v", err)
	}

	event, err := model.NewEventFromOrderedMap(om)
	if err != nil {
		t.Fatalf("failed to parse event: %v", err)
	}

	encoded, err := event.Encode()
	if err != nil {
		t.Fatalf("failed to encode event: %v", err)
	}
	decoded, err := model.DecodeEvent(encoded)
	if err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}

	if decoded.Timestamp != event.Timestamp {
		t.Fatalf("timestamp mismatch: %v != %v", decoded.Timestamp, event.Timestamp)
	}
	if decoded.SpaceID != event.SpaceID {
		t.Fatalf("space_id mismatch: %v != %v", decoded.SpaceID, event.SpaceID)
	}
	if decoded.Publisher != event.Publisher {
		t.Fatalf("publisher mismatch: %v != %v", decoded.Publisher, event.Publisher)
	}
	if decoded.Payload.Ser() != event.Payload.Ser() {
		t.Fatalf("payload mismatch: %v != %v", decoded.Payload, event.Payload)
	}
	if decoded.Spec.InterfaceID != event.Spec.InterfaceID {
		t.Fatalf("interface_id mismatch: %v != %v", decoded.Spec.InterfaceID, event.Spec.InterfaceID)
	}

}

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
			t.Fatalf("failed to parse event: %v", err)
		}

		fmt.Println("hash::")
		fmt.Println(event.Hash())

		exec := model.NewEventExecutionResultFromEvent(event, "success")

		encoded, err := exec.Encode()
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		decoded, err := model.DecodeEventExecutionResult(encoded)
		if err != nil {
			t.Fatalf("DecodeEventExecutionResult failed: %v", err)
		}

		if decoded.Event == nil {
			t.Fatalf("Decoded event is nil")
		}
		if decoded.Event.SpaceID != exec.Event.SpaceID {
			t.Errorf("SpaceID mismatch: got %s, want %s", decoded.Event.SpaceID, exec.Event.SpaceID)
		}
		if decoded.Result != exec.Result {
			t.Errorf("Result mismatch: got %s, want %s", decoded.Result, exec.Result)
		}
		if decoded.EventHash != exec.EventHash {
			t.Errorf("Event hash mismatch: got %s, want %s", decoded.EventHash, exec.EventHash)
		}
		if !decoded.Verify(exec.Hash()) {
			t.Errorf("Hash verification failed")
		}
	})
}
