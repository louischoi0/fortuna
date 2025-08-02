package vm

import (
	"fortuna/core/model"
)

type Operation struct {
	Universe *Universe
	Machine  *StateMachine

	OpCode string
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
		OpCode:   "$write_var$",
		Args:     []interface{}{v},
	}
}

func (o *ABI) ReadVar(address string) *Operation {
	return &Operation{
		Universe: o.Universe,
		Machine:  o.Machine,
		OpCode:   "$read_var$",
		Args:     []interface{}{address},
	}
}

func (o *ABI) WriteMachineState(vector []int64) *Operation {
	return &Operation{
		Universe: o.Universe,
		Machine:  o.Machine,
		OpCode:   "$write_machine_state$",
		Args:     []interface{}{vector},
	}
}
