package storage

import (
	"sync"
	"path/filepath"
)

type FileStorage struct {
	mu		sync.Mutex
	Directory	string
	Files		map[string]*File
}

func NewFileStorage(dir string) *FileStorage {
	return &FileStorage{
		Directory: dir,
		Files: make(map[string]*File),
	}
}

type File struct {
	mu		sync.Mutex
	Storage		*FileStorage
	Name		string
	Path		string
}

func DataRootDir() string {
	return ""
}

func NewFile(storage *FileStorage, name string) *File {
	return &File{
		Storage: storage,
		Path:	filepath.Join(DataRootDir(), storage.Directory, name),
	}
}

func (file *File) AppendFileBytes(data []byte) error {
	f, err := os.OpenFile(file.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func (storage *FileStorage) ListFiles(recursive bool) []string {
	info, err := os.Stat(storage.Dirname)
	if err != nil || !info.IsDir() {
		return []string{}
	}

	items := []string{}
	contents, err := filepath.Glob(filepath.Join(storage.Dirname, "*"))
	if err != nil {
		return []string{}
	}

	for _, item := range contents {
		info, err := os.Stat(item)
		if err != nil {
			continue
		}

		if info.IsDir() && recursive {
			items = append(items, ListFiles(item, recursive)...)
		} else if !info.IsDir() {
			items = append(items, item)
		}
	}

	return items
}
