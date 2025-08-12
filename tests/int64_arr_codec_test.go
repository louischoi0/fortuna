package test

import (
	"fmt"
	"fortuna/util"
	"testing"
)

func TestInt64ArrCodec(t *testing.T) {
	arr := []int64{1, 2, 3, 4, 5}
	encoded := util.EncodeInt64Array(arr)
	decoded, err := util.DecodeInt64Array(encoded)
	if err != nil {
		t.Fatalf("failed to decode int64 array: %v", err)
	}

	fmt.Println(arr)
	fmt.Println(decoded)
}
