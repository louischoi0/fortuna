package service

import (
	"sync"
)

type oracleNode struct {
	address string
	port    int
	spaceID string
}

type NodeGateway struct {
	mu      sync.Mutex
	nodes   []*oracleNode
	nodeMap map[string]*oracleNode
}

func (ng *NodeGateway) AddNode(address string, port int, spaceID string) {
	ng.mu.Lock()
	defer ng.mu.Unlock()

	ng.nodes = append(ng.nodes, &oracleNode{address: address, port: port, spaceID: spaceID})
	ng.nodeMap[spaceID] = &oracleNode{address: address, port: port, spaceID: spaceID}
}

func (ng *NodeGateway) GetNode(spaceID string) *oracleNode {
	ng.mu.Lock()
	defer ng.mu.Unlock()
	return ng.nodeMap[spaceID]
}
