package model

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
