package test

import (
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/structure"

	"testing"
)

func TestEventEncodeDecode(t *testing.T) {
	// Create test data
	publisher := "publisher_000000000000000000000000000000000000000000000000000000"
	spaceID := "space_id_0000000000000000000000000000000000000000000000000000000"
	interfaceID := string(vm.FIC__001)
	kernelVersion := string(vm.BaseV000)

	// Create test params
	params := structure.NewOrderedMap()
	params.Set("test_param", "test_value")
	params.Set("number_param", 123)

	// Create test payload
	payload := structure.NewOrderedMap()
	payload.Set("test_payload", "test_value")
	payload.Set("number_payload", 456)

	// Create EventSpec
	spec := model.NewEventSpec(interfaceID, kernelVersion, params)

	// Create Event
	event, err := model.NewEventRequest(
		publisher,
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

	// Set timestamp for consistent testing
	event.Timestamp = 12345

	// Test Encode
	encoded, err := event.Encode()
	if err != nil {
		t.Fatalf("Failed to encode event: %v", err)
	}

	// Test Decode
	decoded, err := model.DecodeEvent(encoded)
	if err != nil {
		t.Fatalf("Failed to decode event: %v", err)
	}

	if decoded.Timestamp != event.Timestamp {
		t.Fatalf("timestamp mismatch: %v != %v", decoded.Timestamp, event.Timestamp)
	}

	if decoded.Publisher != event.Publisher {
		t.Fatalf("publisher mismatch: %v != %v", decoded.Publisher, event.Publisher)
	}

	if decoded.SpaceID != event.SpaceID {
		t.Fatalf("spaceID mismatch: %v != %v", decoded.SpaceID, event.SpaceID)
	}

	if decoded.Spec.InterfaceID != event.Spec.InterfaceID {
		t.Fatalf("interfaceID mismatch: %v != %v", decoded.Spec.InterfaceID, event.Spec.InterfaceID)
	}

}
