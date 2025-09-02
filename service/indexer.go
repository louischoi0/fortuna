package service

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/rock"
	"fortuna/util"
	"log"
	"sync"

	"github.com/linxGnu/grocksdb"
)

type EventIndex struct {
	EventID string
	SpaceID string
	PageNum uint64
	Length  uint64
	Hash    string
}

type SpaceIndexer struct {
	mu        	sync.Mutex
	Universe  	map[string]*model.Space
	dbs       	map[string]*grocksdb.DB
	indexerDB 	*grocksdb.DB

	eventIndicies 	map[string]map[string]*EventIndex
	events        	map[string]map[string]*model.EventResult

	// EventID <-> SpaceID
	eventSpaceIDMappingCache	map[string]string
	eventCache		map[string]*model.EventResult

	eventCounts	map[string]int64
	//TODO snapshots
}

func (si *SpaceIndexer) GetExecutionInfo(eventID string) (*model.EventResult, error) {
	ex, ok := si.eventCache[eventID]
	if !ok {
		return nil, fmt.Errorf("event %s not found", eventID)
	}
	return ex, nil
}

type MachineStateSnapshot struct {
	MachineID   string
	SpaceID     string
	ActivatedAt int64
	Duration    int64
}

func NewSpaceIndexer(universe map[string]*model.Space, dbs map[string]*grocksdb.DB) *SpaceIndexer {
	db, err := rock.GetDBInstance("indexer.meta")
	if err != nil {
		log.Fatalf("failed to get rocks db: %v", err.Error())
	}

	return &SpaceIndexer{
		Universe:      universe,
		dbs:           dbs,
		indexerDB:     db,
		eventIndicies: make(map[string]map[string]*EventIndex),
		events:        make(map[string]map[string]*model.EventResult),
		eventCounts:   make(map[string]int64),
		eventSpaceIDMappingCache: make(map[string]string),
		eventCache: make(map[string]*model.EventResult),
	}
}

func (si *SpaceIndexer) SelectMachineStateSnapshot(machineID string, spaceID string, activatedAt int64, duration int64) (*MachineStateSnapshot, error) {
	snapshot := &MachineStateSnapshot{
		MachineID:   machineID,
		SpaceID:     spaceID,
		ActivatedAt: activatedAt,
		Duration:    duration,
	}

	return snapshot, nil
}

func (si *SpaceIndexer) ReadEvent(eventID string) (*model.EventResult, error) {
	return nil, nil
}

func (si *SpaceIndexer) GetSpace(spaceID string) (*model.Space, error) {
	space, ok := si.Universe[spaceID]
	if !ok {
		return nil, fmt.Errorf("space %s not found", spaceID)
	}

	return space, nil
}

func (si *SpaceIndexer) ReadPage(spaceID string, pageNum int64) (*model.Page, error) {
	si.mu.Lock()
	defer si.mu.Unlock()

	space, err := si.GetSpace(spaceID)
	if err != nil {
		return nil, err
	}

	page, err := space.ReadPage(uint64(pageNum))
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (si *SpaceIndexer) ScanPages(pages []*model.Page) error {

	for _, page := range pages {
		/**
		lcn, _ := si.GetLastIndexedPageNum(page.SpaceID)

		if page.GetPageNum() <= lcn {
			continue
		}
		**/

		err := si.ScanPage(page)
		if err != nil {
			return err
		}
	}
	return nil
}

func (si *SpaceIndexer) SetLastIndexedPageNum(spaceID string, pageNum uint64) error {
	return rock.SetValue(si.indexerDB, "last_indexed_page", util.EncodeUint64(pageNum))
}

func (si *SpaceIndexer) GetLastIndexedPageNum(spaceID string) (uint64, error) {
	buf, err := rock.GetValue(si.indexerDB, "last_indexed_page")
	if err != nil {
		return 0, err
	}
	
	i, err := util.DecodeUint64(buf)

	if err != nil {
		return 0, err
	}
	
	return i, nil
}


func (si *SpaceIndexer) AddExecution(pageNum uint64, execution *model.EventResult) {
	eventID := execution.Hash()

	spaceID := execution.Event.SpaceID

	if _, ok := si.eventIndicies[spaceID]; !ok{
		si.eventIndicies[spaceID] = make(map[string]*EventIndex)
	}

	if _, ok := si.events[spaceID]; !ok {
		si.events[spaceID] = make(map[string]*model.EventResult)
	}

	si.eventIndicies[spaceID][eventID] = &EventIndex{
		EventID: eventID,
		SpaceID: spaceID,
		PageNum: pageNum,
		Length:  0,
		Hash:    execution.Hash(),
	}

	si.events[spaceID][eventID] = execution
	c, ok := si.eventCounts[spaceID]
	if !ok {
		c = 0
	}

	si.eventCounts[spaceID] = c + 1
	si.eventSpaceIDMappingCache[eventID] = spaceID
	si.eventCache[eventID] = execution
}

func (si *SpaceIndexer) ScanPage(page *model.Page) error {
	log.Printf("indexer scan page %s:%d", page.SpaceID, page.N)

	pageNum := page.GetPageNum()
	defer si.SetLastIndexedPageNum(page.SpaceID, pageNum)

	for _, execution := range page.Executions {
		si.AddExecution(pageNum, execution)
	}

	return nil
}

func (si *SpaceIndexer) ListEvents(spaceID string, count int) ([]*model.EventResult, error) {
	spaceEvents, ok := si.events[spaceID]
	if !ok {
		return nil, fmt.Errorf("space %s events cache map does not exists", spaceID)
	}

	events := make([]*model.EventResult, 0, 100)
	c := 0

	for _, event := range(spaceEvents) {
		events = append(events, event)

		if c > count {
			return events, nil
		}
		c = c + 1
	}
	
	return events, nil
}

func (si *SpaceIndexer) GetEventCounts() map[string]int64 {
	return si.eventCounts
}

