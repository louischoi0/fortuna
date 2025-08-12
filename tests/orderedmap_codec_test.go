package test

import (
	"bytes"
	"fortuna/structure"
	. "fortuna/util"
	"log"
	"reflect"
	"testing"
)

func TestEncodeDecodeSimpleObject(t *testing.T) {
	original := structure.NewOrderedMap()
	original.Set("name", "Alice")
	original.Set("age", int64(30))
	original.Set("tag", "go")
	om := structure.NewOrderedMap()
	om.Set("theme", "dark")
	original.Set("prefs", om)
	original.Set("topic", "topic0")

	encoded, err := EncodeOrderedMap(original)
	if err != nil {
		t.Fatalf("EncodeObject failed: %v", err)
	}

	decoded, err := DecodeOrderedMap(encoded)
	log.Println(original.Keys())
	log.Println(decoded.Keys())

	if err != nil {
		t.Fatalf("DecodeObject failed: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("Decoded object does not match original.\nOriginal: %#v\nDecoded: %#v", original, decoded)
	}

	encoded2, err := EncodeOrderedMap(decoded)
	if err != nil {
		t.Fatalf("Re-Encode failed: %v", err)
	}

	if !bytes.Equal(encoded, encoded2) {
		t.Errorf("Re-encoded bytes are not identical.\nFirst:  %v\nSecond: %v", encoded, encoded2)
	}
}
