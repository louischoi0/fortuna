package core

import (
	"fortuna/core/vm"
	"fortuna/swift"
)

const BASIC_NODE_STATE_COUNT = 256

type Synapse struct {
	Epoch int64

	machines map[string]*vm.StateMachine
	server   *swift.TCPServer
}

func NewBasicSynapse(spaceID string) *Synapse {
	return &Synapse{
		Epoch:    0,
		machines: make(map[string]*vm.StateMachine),
		server:   swift.NewServer(),
	}
}

func (n *Synapse) ResetMachineState(spaceID string) {
	var payload string

	go func(p string) {
		// TODO notify state changes to finalizer
	}(payload)
}

func (n *Synapse) Init(spaceID string) {
	n.ResetMachineState(spaceID)
}
