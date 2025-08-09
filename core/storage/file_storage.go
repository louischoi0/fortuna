package storage

import (
	"os"
	"path/filepath"
	"sync"
)

type FileStorage struct {
	mu        sync.Mutex
	Directory string
	Files     map[string]*File
}

func NewFileStorage(dir string) *FileStorage {
	return &FileStorage{
		Directory: dir,
		Files:     make(map[string]*File),
	}
}

type File struct {
	mu      sync.Mutex
	Storage *FileStorage
	Name    string
	Path    string
	f       *os.File
}

func DataRootDir() string {
	return ""
}

func NewFile(storage *FileStorage, name string) *File {
	file := &File{
		Storage: storage,
		Path:    filepath.Join(DataRootDir(), storage.Directory, name),
	}

	f, err := os.OpenFile(file.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("failed to open file: " + err.Error())
	}
	file.f = f
	return file
}

func (file *File) AppendFileBytes(data []byte) error {
	file.mu.Lock()
	defer file.mu.Unlock()

	if _, err := file.f.Write(data); err != nil {
		return err
	}
	return nil
}

func (file *File) Close() error {
	file.mu.Lock()
	defer file.mu.Unlock()

	return file.f.Close()
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

func (storage *FileStorage) GetFile(name string) *File {
	return storage.Files[name]
}
