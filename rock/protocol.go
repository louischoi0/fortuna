package rock

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/linxGnu/grocksdb"
)

func MetaStoreDataRootDir() string {
	return "data"
}

var (
	dbs      = make(map[string]*grocksdb.DB)
	dbsMutex sync.RWMutex

	dbOnceMap  = make(map[string]*sync.Once)
	onceMapMux sync.Mutex
)

func GetDBInstance(key string) (*grocksdb.DB, error) {
	dbsMutex.RLock()
	db, exists := dbs[key]
	dbsMutex.RUnlock()

	if exists {
		return db, nil
	}

	onceMapMux.Lock()
	once, ok := dbOnceMap[key]
	if !ok {
		once = &sync.Once{}
		dbOnceMap[key] = once
	}
	onceMapMux.Unlock()

	var initErr error

	once.Do(func() {
		opts := grocksdb.NewDefaultOptions()
		opts.SetCreateIfMissing(true)
		os.MkdirAll(MetaStoreDataRootDir(), 0755)

		path := filepath.Join(MetaStoreDataRootDir(), key)

		db, initErr = grocksdb.OpenDb(opts, path)
		if initErr == nil {
			dbsMutex.Lock()
			dbs[key] = db
			dbsMutex.Unlock()
		}
	})

	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize DB for key %s: %v", key, initErr)
	}

	dbsMutex.RLock()
	db, exists = dbs[key]
	dbsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("DB for key %s was not initialized properly", key)
	}

	return db, nil
}

func CloseDB(db *grocksdb.DB) {
	if db != nil {
		db.Close()
	}
}

func ClearDB(db *grocksdb.DB) error {
	wb := grocksdb.NewWriteBatch()
	defer wb.Destroy()

	opts := grocksdb.NewDefaultWriteOptions()
	defer opts.Destroy()

	wb.DeleteRange([]byte{}, []byte{0xFF, 0xFF, 0xFF, 0xFF})
	return db.Write(opts, wb)
}

func SetValue(db *grocksdb.DB, key string, value []byte) error {
	writeOpts := grocksdb.NewDefaultWriteOptions()
	defer writeOpts.Destroy()

	err := db.Put(writeOpts, []byte(key), value)
	if err != nil {
		log.Printf("Failed to write key %s: %v", key, err)
	}
	return err
}

func GetValue(db *grocksdb.DB, key string) ([]byte, error) {
	readOpts := grocksdb.NewDefaultReadOptions()
	defer readOpts.Destroy()

	value, err := db.Get(readOpts, []byte(key))

	if err != nil {
		return nil, err
	}

	defer value.Free()
	return value.Data(), nil
}
