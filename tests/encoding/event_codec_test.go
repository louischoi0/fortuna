package test

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/structure"
	"testing"
)

func TestEncodeDecodeEvent(t *testing.T) {
	ebuf := `{"timestamp": 0, "space_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","publisher": "0000000000000000000000000000000000000000000000000000000000000000","payload":{"a":3},"spec":{"interface_id":"FIC-00-00001","version":"base::v0.0.0","params":{"slot_count":64}},"topic":"","subtopic":"","seperator":"","tag":""}`
	om, err := structure.ParseOrderedMap(ebuf)
	if err != nil {
		t.Fatalf("failed to parse ordered map: %v", err)
	}

	event, err := model.NewEventFromOrderedMap(om)
	if err != nil {
		t.Fatalf("failed to parse event: %v", err)
	}

	encoded, err := event.Encode()
	fmt.Println(err)
	decoded, err := model.DecodeEvent(encoded)
	fmt.Println(err)

	fmt.Println(decoded)
	fmt.Println(decoded.Hash())
	fmt.Println(event.Hash())
}
