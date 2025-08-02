package test

import (
	. "fortuna/core/model"
	"reflect"
	"testing"
)

func TestParseCompactOperation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *OperationRaw
		wantErr  bool
	}{
		{
			name:  "simple string arg",
			input: "*$&0$^hello",
			expected: &OperationRaw{
				OpCode: "$0",
				Args:   []interface{}{"hello"},
			},
		},
		{
			name:  "simple int arg",
			input: "*$&1$!42",
			expected: &OperationRaw{
				OpCode: "$1",
				Args:   []interface{}{int64(42)},
			},
		},
		{
			name:  "nested operation",
			input: "*$&0$^a$*$&1$^b$!2",
			expected: &OperationRaw{
				OpCode: "$0",
				Args: []interface{}{
					"a",
					&OperationRaw{
						OpCode: "$1",
						Args:   []interface{}{"b", int64(2)},
					},
				},
			},
		},
		{
			name:    "missing opcode",
			input:   "*$&",
			wantErr: true,
		},
		{
			name:    "invalid argument",
			input:   "*$&0$xabc",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCompactOperation(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompactOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseCompactOperation() = %+v, expected %+v", got, tt.expected)
			}
		})
	}
}
