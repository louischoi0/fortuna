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
	less  func(a, b QueueContent) bool // comparison function for ordering
}

// NewPriorityQueue creates a new priority queue with a custom less function
func NewPriorityQueue(less func(a, b QueueContent) bool) *PriorityQueue {
	return &PriorityQueue{
		less: less,
	}
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
	return pq.less(pq.items[i], pq.items[j])
}

// Push adds an item to the priority queue
func (pq *PriorityQueue) Push(x interface{}) {
	pq.items = append(pq.items, x.(QueueContent))
	heap.Init(pq) // Re-heapify after adding a new item
}

// Pop removes and returns the highest priority item from the priority queue
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[0 : n-1]
	heap.Init(pq) // Re-heapify after removing an item
	return item
}
