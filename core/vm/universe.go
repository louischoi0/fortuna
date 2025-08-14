package vm

import (
	"fortuna/core/model"
)

type Context struct {
	Height     int64
	SpaceID    string
	GlobalVars map[string]*model.Var
}

func NewContext(spaceID string) *Context {
	return &Context{
		Height:     0,
		SpaceID:    spaceID,
		GlobalVars: make(map[string]*model.Var),
	}
}
