package model

type VarType int

const (
	VarTypeInt    VarType = 0
	VarTypeString VarType = 1
	VarTypeBool   VarType = 2
	VarTypeVector VarType = 3
)

type Var struct {
	SpaceID string
	Value   interface{}
	Size    uint64
	Type    VarType

	Address string
}

func (v *Var) GetValue() interface{} {
	return v.Value
}

func (v *Var) GetSize() uint64 {
	return v.Size
}
