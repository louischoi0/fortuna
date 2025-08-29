package service

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/rock"
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
	mu        sync.Mutex
	Universe  map[string]*model.Space
	dbs       map[string]*grocksdb.DB
	indexerDB *grocksdb.DB

	eventIndicies map[string]map[string]*EventIndex
	events        map[string]map[string]*model.EventResult

	//TODO snapshots
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
		err := si.ScanPage(page)
		if err != nil {
			return err
		}
	}

	return nil
}

func (si *SpaceIndexer) ScanPage(page *model.Page) error {
	spaceID := page.SpaceID
	pageNum := page.GetPageNum()

	for _, execution := range page.Executions {
		eventID := execution.Event.ID
		si.eventIndicies[spaceID][eventID] = &EventIndex{
			EventID: eventID,
			SpaceID: spaceID,
			PageNum: pageNum,
			Length:  0,
			Hash:    execution.Hash(),
		}
		si.events[spaceID][eventID] = execution
	}

	return nil
}
