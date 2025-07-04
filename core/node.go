package core

import (
	"fortuna/swift"
	"fortuna/core/vm"
)

const BASIC_NODE_STATE_COUNT = 256

type Synapse struct {
	Epoch		int64
	machines 	map[string]*vm.StateMachine
	server		*swift.TCPServer
}

func NewBasicSynapse(spaceID string) *Synapse {
	return &Synapse{
		Epoch: 0,
		machine: make(map[string]*vm.StateMachine),
		server: swift.NewServer(),
	}
}

func (n *Synapse) ResetMachineState(workspaceID string) {
	var payload string

	n.State, payload = n.machine.GetNodeState(n.StateCount)
	n.Epoch++

	go func(p string) {
		// TODO notify state changes to finalizer
	}(payload)
}

func (n *Synapse) Init() {
	n.ResetState()

}
