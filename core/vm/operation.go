package vm

import (
	"fmt"
	"fortuna/core/model"
	"strconv"
	"strings"
)

type Operation struct {
	Universe *Universe
	Machine  *StateMachine

	OpCode string
	OpName string
	Args   []interface{}
}

type ABI struct {
	Universe *Universe
	Machine  *StateMachine
}

func (o *Operation) WriteVar(v *model.Var) *Operation {
	return &Operation{
		Universe: o.Universe,
		Machine:  o.Machine,
		OpCode:   "x0",
		OpName:   "write_var",
		Args:     []interface{}{v},
	}
}

func (o *ABI) ReadVar(address string) *Operation {
	return &Operation{
		Universe: o.Universe,
		Machine:  o.Machine,
		OpCode:   "x1",
		OpName:   "read_var",
		Args:     []interface{}{address},
	}
}

func (o *ABI) WriteMachineState(vector []int64) *Operation {
	return &Operation{
		Universe: o.Universe,
		Machine:  o.Machine,

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

	sb.WriteByte(model.OP_SYMBOL)

	sb.WriteString(op.OpCode)

	for _, arg := range op.Args {
		sb.WriteByte(model.ARG_SYMBOL)
		switch v := arg.(type) {
		case string:
			sb.WriteByte(model.STR_SYMBOL)
			sb.WriteString(v)
		case int64:
			sb.WriteByte(model.INT_SYMBOL)
			sb.WriteString(strconv.FormatInt(v, 10))
		case *Operation:
			if err := serializeOperation(sb, v); err != nil {
				return fmt.Errorf("nested operation serialize failed: %v", err)
			}
		default:
			return fmt.Errorf("unsupported argument type: %T", v)
		}
	}
	return nil
}
