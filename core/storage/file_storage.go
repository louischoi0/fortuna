package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"log"
	"sync"
)

type FileStorage struct {
	mu        sync.Mutex
	Directory string
	files	  map[string]*File
}

func NewFileStorage(dir string) *FileStorage {
	os.MkdirAll(dir, 0755)

	return &FileStorage{
		Directory: dir,
		files: make(map[string]*File),
	}
}

func (fs *FileStorage) Clear() error {
	for fn := range(fs.files) {
		err := fs.files[fn].Clear()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}
	return nil
}

type File struct {
	mu      sync.Mutex
	Storage *FileStorage
	Name    string
	Path    string
	f	*os.File
}

func DataRootDir() string {
	return "data"
}

func NewFile(storage *FileStorage, name string) *File {
	path := filepath.Join(storage.Directory, name)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 644)
	if err != nil {
		log.Fatal(err.Error())
	}

	file := &File{
		Storage: storage,
		Path: path,
		f: f,
	}
	return file
}

func (file *File) Clear() error {
	log.Printf("clear storage file %s", file.Path)
	return file.f.Close()
}

func (file *File) AppendFileBytes(data []byte) error {
	file.mu.Lock()
	defer file.mu.Unlock()

	if _, err := file.f.Write(data); err != nil {
		return err
	}

	return nil
}

func (file *File) GetSize() int64 {
	file.mu.Lock()
	defer file.mu.Unlock()
	
	info, err := file.f.Stat()
	if err != nil {
		return 0
	}

	return info.Size()
}

func (storage *FileStorage) ListFiles(recursive bool) []string {
	info, err := os.Stat(storage.Directory)
	if err != nil || !info.IsDir() {
		return []string{}
	}

	items := []string{}
	contents, err := filepath.Glob(filepath.Join(storage.Directory, "*"))
	if err != nil {
		return []string{}
	}

	for _, item := range contents {
		info, err := os.Stat(item)
		if err != nil {
			continue
		}

		if info.IsDir() && recursive {
			items = append(items, storage.ListFiles(recursive)...)
		} else if !info.IsDir() {
			items = append(items, item)
		}
	}

	return items
}

func (file *File) OffsetRead(offset int64, size int64) ([]byte, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	_, err := file.f.Seek(offset, os.SEEK_SET)
	if err != nil {
		return nil, fmt.Errorf("failed to seek file: %v", err)
	}

	buffer := make([]byte, size)
	n, err := file.f.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("failed to read from file: %v", err)
	}

	return buffer[:n], nil
}
