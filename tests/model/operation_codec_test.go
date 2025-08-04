package test

import (
	"fortuna/core/model"
	"testing"
)

func TestOperationCodec(t *testing.T) {
	v := model.NewStateVector([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	o := model.NewOperation("x0", "write_var", []interface{}{v})
	op_str, err := model.SerializeCompactOperations([]*model.Operation{o})
	if err != nil {
		t.Fatalf("failed to serialize operation: %v", err)
	}

	op, err := model.ParseCompactOperation(op_str)
	if err != nil {
		t.Fatalf("failed to parse operation: %v", err)
	}

	t.Logf("operation: %s", op)
}
