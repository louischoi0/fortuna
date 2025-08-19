package component

import (
	"fortuna/core/model"
	"sort"
	"sync"
	"fmt"
)

const PAGE_BUFFER_CHAN_SIZE = 256

type PageBuffer struct {
	mu 	sync.RWMutex
	buffer 	map[uint64]*model.Page
	pq	[]uint64

	insig	chan *model.Page
	wait	bool
}


func NewPageBuffer() *PageBuffer {
	return &PageBuffer{
		buffer: make(map[uint64]*model.Page),
		pq: make([]uint64, 0, 256),
		insig: make(chan *model.Page, 256*8),
	}
}

func (pb *PageBuffer) AddPage(page *model.Page) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.insig <- page

	page_num := page.N

	if _, exists := pb.buffer[page_num]; exists {
		return fmt.Errorf("page %v already exists in buffer", page_num)
	}

	pb.buffer[page_num] = page	
	pb.addpq(page_num)

	return nil
}

func (pb *PageBuffer) RemovePage(page *model.Page) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if _, exists := pb.buffer[page.N]; exists {
		delete(pb.buffer, page.N)
		pb.removepq(page.N)
	}
	return fmt.Errorf("page %v does not exists in buffer", page.N)
}


func (pb *PageBuffer) Clear() {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	
	pb.buffer = make(map[uint64]*model.Page)
	pb.pq = make([]uint64, 0, 256)
}

func (pb *PageBuffer) addpq(n uint64) error {
	pb.pq = append(pb.pq, n)
	pb.sortpq()
	return nil
}

func (pb *PageBuffer) removepq(n uint64) error {
	for i, pn := range pb.pq {
		if pn == n {
			pb.pq = append(pb.pq[:i], pb.pq[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("page %s not found in priority queue", n)
}


func (pb *PageBuffer) sortpq() {
	sort.Slice(pb.pq, func(i, j int) bool {
		return pb.pq[i] < pb.pq[j]
	})
}

func (pb *PageBuffer) Size() int {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	
	return len(pb.buffer)
}

func (pb *PageBuffer) Pop() *model.Page {
	pb.mu.Lock()
	defer pb.mu.Unlock()	
	
	if len(pb.buffer) == 0 {
		return nil
	}

	num := pb.pq[0]	
	page := pb.buffer[num]
	
	delete(pb.buffer, num)
	pb.pq = pb.pq[1:]

	return page
}

func (pb *PageBuffer) First() (uint64, error) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if len(pb.buffer) == 0 {
		return 0, fmt.Errorf("Queue is empty")
	}

	return pb.pq[0], nil
}

func (pb *PageBuffer) WaitFor(pageNum uint64) *model.Page {
    for {
        pb.mu.Lock()
        if len(pb.pq) > 0 && pb.pq[0] == pageNum {
		return pb.buffer[pageNum]
        }
        pb.mu.Unlock()

        <-pb.insig
    }
}
