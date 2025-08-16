package vm

import (
	"fortuna/core/model"
)

type Context struct {
	SpaceID    string
	GlobalVars map[string]*model.Var
}

func NewContext(spaceID string) *Context {
	return &Context{
		SpaceID:    spaceID,
		GlobalVars: make(map[string]*model.Var),
	}
}
