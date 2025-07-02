package core

const BASIC_NODE_STATE_COUNT = 256

type Node struct {
	Epoch		int64
	State 		[]float64
	StateHash	string	
	StateCount	int64

	machine		*StateMachine
}


func NewBasicNode() *Node {
	return &Node{
		Epoch: 0,
		State: make([]float64, BASIC_NODE_STATE_COUNT),
		StateCount: 256,
	}
}

func (n *Node) ResetState() {
	n.State = n.machine.StateKernel.GenVector(n.StateCount)
	n.Epoch++
}

func (n *Node) Init() {
	n.ResetState()
}
