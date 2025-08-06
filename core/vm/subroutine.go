package vm

import (
	"fortuna/core/model"
	"fmt"
)

type SubroutineParam struct {
	Name		string
	Type		model.VarType
	Required	bool
	MaxVal		*int64
	Value 		*model.Var
}

type Subroutine struct {
	Params		map[string]*SubroutineParam
	Operations	[]*model.Operation
}

func (sr *Subroutine) AddParam(name string, paramType model.VarType, required bool, maxVal *int64) {
	param := &SubroutineParam{
		Name: name,
		Type: paramType,
		Required: required,
		MaxVal:	maxVal,
		Value: nil,
	}
	sr.Params[name] = param
}

func (sr *Subroutine) SetParamValue(name string, value *model.Var) error {
	param, ok := sr.Params[name]
	if !ok {
		return fmt.Errorf("unknown param for subroutine: %v", name)
	}

	param.Value = value
	return nil
}

func (sr *Subroutine) AddOperation(op *model.Operation) {
	sr.Operations = append(sr.Operations, op)
}

func NewLogMachineStateSubroutine(machineID string, stateVector *model.StateVector) *Subroutine {
	subroutine := &Subroutine{}
	abi := &ABI{}

	write_op := abi.WriteVar(machineID, stateVector.Var())
	subroutine.AddOperation(write_op)

	return subroutine
}

