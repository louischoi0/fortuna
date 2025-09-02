package rock

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/linxGnu/grocksdb"
)

var MetastoreRootDir = "data"
var LOGGING = false

func MetaStoreDataRootDir() string {
	return MetastoreRootDir
}

func SetMetastoreDataRootDir(dir string) {
	MetastoreRootDir = dir
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
		log.Fatalf("failed to initialize DB for key %s: %v", key, initErr)
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
	if LOGGING {
		log.Printf("rock set - key: %s, size: %v", key, len(value))
	}

	writeOpts := grocksdb.NewDefaultWriteOptions()
	defer writeOpts.Destroy()

	err := db.Put(writeOpts, []byte(key), value)
	if err != nil {
		log.Fatalf("Failed to write key %s: %v", key, err)
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

	data := value.Data()
	defer value.Free()

	ret := make([]byte, len(data))
	copy(ret, data)

	return ret, nil
}

func Sacn(db *grocksdb.DB, prefix string) (map[string][]byte, error) {
	readOpts := grocksdb.NewDefaultReadOptions()
	defer readOpts.Destroy()

	it := db.NewIterator(readOpts)
	defer it.Close()

	results := make(map[string][]byte)

	it.Seek([]byte(prefix))
	for ; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()

		if !bytes.HasPrefix(key.Data(), []byte(prefix)) {
			key.Free()
			value.Free()
			break
		}

		k := make([]byte, len(key.Data()))
		copy(k, key.Data())

		v := make([]byte, len(value.Data()))
		copy(v, value.Data())

		results[string(k)] = v

		key.Free()
		value.Free()
	}

	if err := it.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func ScanC[T any](db *grocksdb.DB, prefix string, cb func(key string, value []byte) T) ([]T, error) {
	readOpts := grocksdb.NewDefaultReadOptions()
	defer readOpts.Destroy()

	it := db.NewIterator(readOpts)
	defer it.Close()

	var results []T

	it.Seek([]byte(prefix))
	for ; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()

		if !bytes.HasPrefix(key.Data(), []byte(prefix)) {
			key.Free()
			value.Free()
			break
		}

		k := string(key.Data())
		v := make([]byte, len(value.Data()))
		copy(v, value.Data())

		results = append(results, cb(k, v))

		key.Free()
		value.Free()
	}

	if err := it.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
