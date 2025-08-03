package test

import (
	"encoding/json"
	"fmt"
	. "fortuna/core/model"
	"testing"
)

func TestParseCompactOperation(t *testing.T) {
	tests := []struct {
		input    string
		wantCode string
		wantErr  bool
	}{
		{"*x00$^hello$!123", "x00", false},
		{"*x01$!42$^str$*x02$^key$!1", "x01", false},
		{"*x00$^hello$!badnum", "", true},
		{"x00$^hello", "", true},
		{"*x00^hello", "", true},
		{"*x00$^", "", true},
		{"*x00$!", "", true},
	}

	for _, tt := range tests {
		op, err := ParseCompactOperation(tt.input)
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

func TestParseCompactOperations(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []*OperationRaw
		wantErr bool
	}{
		{
			name:  "two simple operations",
			input: "*x00$^hello$!123*x01$^world$!456",
			want: []*OperationRaw{
				{
					OpCode: "x00",
					Args:   []interface{}{"hello", int64(123)},
				},
				{
					OpCode: "x01",
					Args:   []interface{}{"world", int64(456)},
				},
			},
			wantErr: false,
		},
		{
			name:  "operation with nested operation",
			input: "*x00$*x01$^nested$!1$!2$^outer",
			want: []*OperationRaw{
				{
					OpCode: "x00",
					Args: []interface{}{
						&OperationRaw{
							OpCode: "x01",
							Args:   []interface{}{"nested", int64(1), int64(2)},
						},
						"outer",
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "operation with whitespace separators",
			input: "*x00$!1*x00$!2*x00$!3",
			want: []*OperationRaw{
				{OpCode: "x00", Args: []interface{}{int64(1)}},
				{OpCode: "x00", Args: []interface{}{int64(2)}},
				{OpCode: "x00", Args: []interface{}{int64(3)}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseCompactOperations(tt.input)

			fmt.Printf("got: %v", len(got))
			fmt.Printf("want: %v", len(tt.want))
		})
	}
}
