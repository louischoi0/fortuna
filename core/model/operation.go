package model

import (
	"fmt"
	"strconv"
	"strings"
)

type Operation struct {
	OpCode string
	OpName string
	Args   []interface{}
}

func NewOperation(opCode string, opName string, args []interface{}) *Operation {
	return &Operation{
		OpCode: opCode,
		OpName: opName,
		Args:   args,
	}
}

type ABI struct {
}

func (o *ABI) WriteVar(v *Var) *Operation {
	return &Operation{
		OpCode: "x0",
		OpName: "write_var",
		Args:   []interface{}{v},
	}
}

func (o *ABI) ReadVar(address string) *Operation {
	return &Operation{
		OpCode: "x1",
		OpName: "read_var",
		Args:   []interface{}{address},
	}
}

func (o *ABI) WriteMachineState(vector *StateVector) *Operation {
	return &Operation{
		OpCode: "x2",
		OpName: "write_machine_state",
		Args:   []interface{}{vector},
	}
}

func SerializeCompactOperations(ops []*Operation) (string, error) {
	var sb strings.Builder
	for _, op := range ops {
		if err := serializeOperation(&sb, op); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

func serializeOperation(sb *strings.Builder, op *Operation) error {
	if op == nil {
		return fmt.Errorf("nil operation")
	}

	sb.WriteByte(OP_SYMBOL)
	sb.WriteString(op.OpCode)

	for _, arg := range op.Args {
		sb.WriteByte(ARG_MARK)
		switch v := arg.(type) {
		case string:
			sb.WriteByte(STR_SYMBOL)
			sb.WriteString(v)
			sb.WriteByte(STR_END_SYMBOL)
		case int64:
			sb.WriteByte(INT_SYMBOL)
			sb.WriteString(strconv.FormatInt(v, 10))
			sb.WriteByte(INT_END_SYMBOL)
		case *StateVector:
			sb.WriteByte(VEC_SYMBOL)
			sb.Write(v.Encode())
			sb.WriteByte(VEC_END_SYMBOL)
		case *Operation:
			if err := serializeOperation(sb, v); err != nil {
				return fmt.Errorf("nested operation serialize failed: %v", err)
			}
		default:
			return fmt.Errorf("unsupported argument type: %T", v)
		}
	}
	sb.WriteByte(OP_END_SYMBOL)
	return nil
}
