package service

import (
	"fortuna/core/model"
	"fortuna/rock"
	"log"
	"sync"

	"github.com/linxGnu/grocksdb"
)

type SpaceIndexer struct {
	mu        sync.Mutex
	Universe  map[string]*model.Space
	dbs       map[string]*grocksdb.DB
	indexerDB *grocksdb.DB
}

type EventIndex struct {
	EventID string
	SpaceID string
	PageNum int64
	Length  int64
	Hash    string
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
		Universe:  universe,
		dbs:       dbs,
		indexerDB: db,
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
