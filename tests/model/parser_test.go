package test

import (
	"encoding/json"
	"fmt"
	. "fortuna/core/model"
	"testing"
)

func TestParseCompactOperation(t *testing.T) {
	vector := NewStateVector([]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	buf := vector.Encode()

	tests := []struct {
		input    string
		wantCode string
		wantErr  bool
	}{
		{`*x00$"hello"$%` + string(buf) + "%;", "x00", false},
	}

	for _, tt := range tests {
		fmt.Println(tt.input)

		op, err := ParseCompactOperation([]byte(tt.input))
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseCompactOperation(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if err == nil && op.OpCode != tt.wantCode {
			t.Errorf("got OpCode = %s, want %s", op.OpCode, tt.wantCode)
		}
		if err == nil {
			out, _ := json.MarshalIndent(op, "", "  ")
			t.Logf("Parsed: %s", out)
		}
	}
}
