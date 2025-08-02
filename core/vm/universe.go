package vm

import (
	"fortuna/core/model"
)

type Universe struct {
	Height     int64
	SpaceID    string
	GlobalVars map[string]*model.Var
}

func NewUniverse(spaceID string) *Universe {
	return &Universe{
		Height:     0,
		SpaceID:    spaceID,
		GlobalVars: make(map[string]*model.Var),
	}
}
