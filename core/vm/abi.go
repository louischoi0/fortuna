package vm

import (
	"fortuna/core/model"
)

type ABI struct {}

func (o *ABI) WriteVar(k, v interface{}) *model.Operation {
	return &model.Operation{
		OpCode: "x0",
		OpName: "write_var",
		Args:   []interface{}{k, v},
	}
}

func (o *ABI) ReadVar(address interface{}) *model.Operation {
	return &model.Operation{
		OpCode: "x1",
		OpName: "read_var",
		Args:   []interface{}{address},
	}
}

func (o *ABI) WriteMachineState(vector interface{}) *model.Operation {
	return &model.Operation{
		OpCode: "x2",
		OpName: "write_machine_state",
		Args:   []interface{}{vector},
	}
}

func (o *ABI) Param(name interface{}) *model.Operation {
	return &model.Operation{
		OpCode: "x3",
		OpName: "param",
		Args:   []interface{}{name},
	}
}

