package model

import (
	"time"
)

type Chain struct {
	SpaceID 	string
	Blocks  	[]*Block

	LastHeight     	int64
	LastAppendedAt 	time.Time
	LastVerifiedAt 	time.Time
}
