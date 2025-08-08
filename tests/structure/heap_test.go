package test

import (
	"container/heap"
	"fmt"
	. "fortuna/structure"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue(func(a, b QueueContent) bool {
		return a.GetPriority() < b.GetPriority() // Lower priority number means higher priority
	})

	// Use the heap interface to manage the priority queue
	heap.Push(pq, &ExampleQueueContent{priority: 3, name: "Low Priority"})
	heap.Push(pq, &ExampleQueueContent{priority: 1, name: "High Priority"})
	heap.Push(pq, &ExampleQueueContent{priority: 2, name: "Medium Priority"})

	// Pop and print the highest priority item
	item := heap.Pop(pq).(*ExampleQueueContent)
	fmt.Println("Popped item:", item.name) // Should print "High Priority"
}
