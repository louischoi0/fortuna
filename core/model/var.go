package model

import (
	"log"
)

type VarType int

type VarValue interface {
	Var() *Var
	Encode() []byte
}

type Var struct {
	Value   VarValue
	Type    VarType
}

const (
	VarTypeInt    VarType = 0
	VarTypeString VarType = 1
	VarTypeBool   VarType = 2
	VarTypeVector VarType = 3
)

func (v *Var) GetValue() interface{} {
	return v.Value
}

func (v *Var) Vec() *StateVector {
	if v.Type != VarTypeVector {
		log.Fatalf("call Vec() on var with %d", v.Type)
	}
	sv, ok := (v.Value).(*StateVector)
	if !ok {
		log.Fatalf("var.Value conversion to state vector failed")
	}
	return sv
}

