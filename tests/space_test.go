package test

import (
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/structure"
	"log"
	"testing"
)

func NewTestEvent() *model.Event {
	publisher := "publisher_000000000000000000000000000000000000000000000000000000"
	spaceID := "space_id_0000000000000000000000000000000000000000000000000000000"
	interfaceID := string(vm.FIC__001)
	kernelVersion := string(vm.BaseV000)

	params := structure.NewOrderedMap()
	params.Set("test_param", "test_value")
	params.Set("number_param", 123)

	payload := structure.NewOrderedMap()
	payload.Set("test_payload", "test_value")
	payload.Set("number_payload", 456)

	spec := model.NewEventSpec(interfaceID, kernelVersion, params)

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
		log.Fatalf("failed to create event: %v", err)
	}

	return event
}

func TestSpace(t *testing.T) {
	space := model.NewSpace("space_id_0000000000000000000000000000000000000000000000000000000")
	event := NewTestEvent()

	space.CurrentPage.AppendEventExecution(model.NewEventResultFromEvent(event, "success"))
	space.CurrentPage.UpdateTransactionRoot()
	space.CurrentPage.UpdateExecutionRoot()

	page, err := space.CurrentPage.Encode()
	if err != nil {
		t.Fatalf("failed to encode page: %v", err)
	}

	decoded_page, err := model.DecodePage(page)
	if err != nil {
		t.Fatalf("failed to decode page: %v", err)
	}

	if decoded_page.Timestamp != space.CurrentPage.Timestamp {
		t.Fatalf("timestamp mismatch: %v != %v", decoded_page.Timestamp, space.CurrentPage.Timestamp)
	}

	if decoded_page.TransactionRootHash != space.CurrentPage.TransactionRootHash {
		t.Fatalf("transaction root hash mismatch: %v != %v", decoded_page.TransactionRootHash, space.CurrentPage.TransactionRootHash)
	}

}
