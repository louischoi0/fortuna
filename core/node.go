package core

import (
	"fortuna/swift"
)

const BASIC_NODE_STATE_COUNT = 256

type Node struct {
	Epoch		int64
	State 		[]float64
	StateHash	string	
	StateCount	int64

	machine		*StateMachine
	streamer	interface{}

	server		*swift.TCPServer
}

func NewBasicNode(spaceID string) *Node {
	return &Node{
		Epoch: 0,
		machine: NewBasicStateMachine(spaceID),
		State: make([]float64, BASIC_NODE_STATE_COUNT),
		server: swift.NewServer(),
		StateCount: 256,
	}
}

func (n *Node) ResetState() {
	var payload string

	n.State, payload = n.machine.GetNodeState(n.StateCount)
	n.Epoch++

	go func(p string) {
		// TODO notify state changes to finalizer
	}(payload)
}

func (n *Node) Init() {
	n.ResetState()

}
