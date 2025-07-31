package model

import (
	"time"
)


type SpaceChain struct {
	SpaceID		string

	Blocks		[]*EventBlock

	LastHeight	int64
	LastAppendedAt	time.Time
	LastVerifiedAt	time.Time
}

