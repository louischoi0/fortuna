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

	fmt.Println(decoded)
	fmt.Println(decoded.Hash())
	fmt.Println(event.Hash())
	fmt.Println(event.Timestamp)
}
