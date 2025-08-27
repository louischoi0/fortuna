package service 

import (
	"fortuna/core/model"
	"fortuna/rock"
	"sync"
	"log"
	"github.com/linxGnu/grocksdb"
)

type SpaceIndexer struct {
	mu		sync.Mutex	
	Universe	map[string]*model.Space
	dbs             map[string]*grocksdb.DB
	indexerDB	*grocksdb.DB
}

func NewSpaceIndexer(universe map[string]*model.Space, dbs map[string]*grocksdb.DB) *SpaceIndexer {
        db, err := rock.GetDBInstance("indexer.meta")
        if err != nil {
                log.Fatalf("failed to get rocks db: %v", err.Error())
        }
	
	return &SpaceIndexer{
		Universe: universe,
		dbs: dbs,
		indexerDB: db,
	}
}





