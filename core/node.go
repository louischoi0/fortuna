package core

import (
	"math/rand"
	"crypto/sha256"
	"encoding/hex"
	"hash/fnv"
	"time"
)

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
		State: make([]float64, stateCount),
		StateCount: 256,
		StateKernel: &BasicStateKernel{},
		HashKernel: &BasicHashKernel{},
	}
}

func (n *Node) ResetState() {
	n.State = n.StateKernel.GenVector(n.StateCount)
}

func (n *Node) Init() {
	n.ResetState()
}

