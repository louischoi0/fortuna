package structure

import (
	"container/heap"
)

// QueueContent interface that requires the GetPriority method
type QueueContent interface {
	GetPriority() int
}

// PriorityQueue defines the structure for the priority queue
type PriorityQueue struct {
	items []QueueContent
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{}
}

// Len returns the number of items in the priority queue
func (pq *PriorityQueue) Len() int {
	return len(pq.items)
}

// Swap swaps the items with indexes i and j
func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Less reports whether the element with index i should sort before the element with index j
func (pq *PriorityQueue) Less(i, j int) bool {
	// Compare the priority of the two items using GetPriority
	return pq.items[i].GetPriority() < pq.items[j].GetPriority()
}

// Push adds an item to the priority queue
func (pq *PriorityQueue) Push(x interface{}) {
	pq.items = append(pq.items, x.(QueueContent))
	// Re-heapify after adding a new item
	heap.Push(pq, x)
}

// Pop removes and returns the highest priority item from the priority queue
func (pq *PriorityQueue) Pop() interface{} {
	// Pop the item and maintain heap order
	return heap.Pop(pq)
}
