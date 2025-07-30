package util

import (
	"bytes"
	. "fortuna/core/util"
	"reflect"
	"testing"
)

func equalBytes(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func equalInt64s(a, b []int64) bool {
	return reflect.DeepEqual(a, b)
}

func TestEncodeInt64Array_EndianAndContent(t *testing.T) {
	in := []int64{1, -2}
	want := []byte{
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	}
	got := EncodeInt64Array(in)
	if !equalBytes(got, want) {
		t.Fatalf("encode bytes mismatch:\n  want: %v\n  got:  %v", want, got)
	}
}

func TestEncodeDecode_Empty(t *testing.T) {
	var in []int64
	enc := EncodeInt64Array(in)
	if len(enc) != 0 {
		t.Fatalf("encode empty should be empty, got %d bytes", len(enc))
	}
	dec := DecodeInt64Array(enc)
	if len(dec) != 0 {
		t.Fatalf("decode empty should be empty slice, got %v", dec)
	}
}

func TestDecodeInt64Array_InvalidLength(t *testing.T) {
	data := append(
		[]byte{1, 0, 0, 0, 0, 0, 0, 0},
		0x99,
	)
	_ = DecodeInt64Array(data)
}
