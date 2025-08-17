package component

import (
	"fortuna/core/model"
)

type Verifier struct {
	Space *model.Space
}

func NewVerifier(space *model.Space) *Verifier {
	return &Verifier{
		Space: space,
	}
}
