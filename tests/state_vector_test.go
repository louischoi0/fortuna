package test

import (
	"fmt"
	"fortuna/core/model"
	"testing"
)

func TestStateVector(t *testing.T) {
	vector := model.NewStateVector([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	encoded := vector.Encode()
	decoded, err := model.DecodeStateVector(encoded)
	if err != nil {
		t.Fatalf("failed to decode state vector: %v", err)
	}

	fmt.Println(vector)
	fmt.Println(decoded)
}
