package model

import (
	"bytes"
	"fmt"
	"strconv"
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

func SerializeCompactOperations(ops []*Operation) ([]byte, error) {
	var sb bytes.Buffer

	for _, op := range ops {
		if err := serializeOperation(&sb, op); err != nil {
			return nil, err
		}
	}
	return sb.Bytes(), nil
}

func serializeVar(sb *bytes.Buffer, arg interface{}) error {
	switch v := arg.(type) {
	case string:
		sb.WriteByte(STR_SYMBOL)
		sb.WriteString(v)
		sb.WriteByte(STR_END_SYMBOL)
	case int:
		sb.WriteByte(INT_SYMBOL)
		// TODO
		sb.WriteString(strconv.FormatInt(int64(v), 10))
		sb.WriteByte(INT_END_SYMBOL)
	case int64:
		sb.WriteByte(INT_SYMBOL)
		// TODO
		sb.WriteString(strconv.FormatInt(v, 10))
		sb.WriteByte(INT_END_SYMBOL)
	case *StateVector:
		sb.WriteByte(VEC_SYMBOL)
		sb.Write(v.Encode())
		sb.WriteByte(VEC_END_SYMBOL)
	case *Var:
		return serializeVar(sb, v.Value)
	case *Operation:
		if err := serializeOperation(sb, v); err != nil {
			return fmt.Errorf("nested operation serialize failed: %v", err)
		}
	default:
		return fmt.Errorf("unsupported argument type: %T", v)
	}
	return nil
}

func serializeOperation(sb *bytes.Buffer, op *Operation) error {
	if op == nil {
		return fmt.Errorf("nil operation")
	}

	sb.WriteByte(OP_SYMBOL)
	sb.WriteString(op.OpCode)

	for _, arg := range op.Args {
		sb.WriteByte(ARG_MARK)
		if err := serializeVar(sb, arg); err != nil {
			return err
		}
	}
	sb.WriteByte(OP_END_SYMBOL)
	return nil
}
