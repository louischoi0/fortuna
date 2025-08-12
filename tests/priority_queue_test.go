package test

import (
	. "fortuna/structure"
	"testing"
)

// QueueContent implementation for testing
type Item struct {
	value    string
	priority int
}

// GetPriority returns the priority of the item
func (i *Item) GetPriority() int {
	return i.priority
}

// Custom comparison function for priority queue
func priorityComparison(a, b QueueContent) bool {
	return a.GetPriority() < b.GetPriority() // Lower priority value comes first
}

func TestPriorityQueue(t *testing.T) {
	// Create a new priority queue with a custom comparison function
	pq := NewPriorityQueue()

	// Create test data items with varying priorities
	item1 := &Item{value: "item1", priority: 3}
	item2 := &Item{value: "item2", priority: 1}
	item3 := &Item{value: "item3", priority: 2}

	// Push items into the priority queue
	pq.Push(item1)
	pq.Push(item2)
	pq.Push(item3)

	// Verify the length of the queue
	if pq.Len() != 3 {
		t.Errorf("Expected length 3, got %d", pq.Len())
	}

	// Pop the highest priority item (lowest priority number)
	poppedItem := pq.Pop().(*Item)
	if poppedItem.value != "item2" {
		t.Errorf("Expected item2, got %s", poppedItem.value)
	}

	// After popping item2, verify that the length is now 2
	if pq.Len() != 2 {
		t.Errorf("Expected length 2, got %d", pq.Len())
	}

	// Pop the next highest priority item
	poppedItem = pq.Pop().(*Item)
	if poppedItem.value != "item3" {
		t.Errorf("Expected item3, got %s", poppedItem.value)
	}

	// Pop the last item
	poppedItem = pq.Pop().(*Item)
	if poppedItem.value != "item1" {
		t.Errorf("Expected item1, got %s", poppedItem.value)
	}

	// Verify the queue is empty after popping all items
	if pq.Len() != 0 {
		t.Errorf("Expected length 0, got %d", pq.Len())
	}
}
