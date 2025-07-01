package core

import (
	"fortuna/structure"
)

type Identity struct {
}

type EventProtocol struct {
	ProtocolID	int64
	ProtocolName	string
}

type Event struct {
	Publisher 	Identity
	Protocol	EventProtocol
	Payload		*structure.OrderedMap
}

