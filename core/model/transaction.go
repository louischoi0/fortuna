package model

import (
	"fortuna/structure"
)

type Transaction struct {
	SpaceID string `json:"space_id"`
	From    string `json:"from"`

	Params    *structure.OrderedMap `json:"params"`
	Signature string                `json:"signature"`
	Xpub      string                `json:"xpub"`

	Executed bool
	// Operations 		[]*Operation
}
