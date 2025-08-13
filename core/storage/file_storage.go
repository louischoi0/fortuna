package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type FileStorage struct {
	mu        sync.Mutex
	Directory string
	files     map[string]*File
}

func NewFileStorage(dir string) *FileStorage {
	os.MkdirAll(dir, 0755)

	return &FileStorage{
		Directory: dir,
		files:     make(map[string]*File),
	}
}

func (fs *FileStorage) GetOrOpenFile(name string) *File {
	if file, ok := fs.files[name]; ok {
		return file
	}

	file := NewFile(fs, name)
	fs.files[name] = file
	return file
}

func (fs *FileStorage) Close() error {
	for fn := range fs.files {
		err := fs.files[fn].Close()
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
	f       *os.File
	removed bool
}

func DataRootDir() string {
	return "data"
}

func NewFile(storage *FileStorage, name string) *File {
	path := filepath.Join(storage.Directory, name)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}

	file := &File{
		Storage: storage,
		Path:    path,
		removed: false,
		f:       f,
	}
	return file
}

func (file *File) Delete() error {
	file.f.Close()
	err := os.Remove(file.Path)
	if err != nil {
		return err
	}
	file.removed = true
	return nil
}

func (file *File) Close() error {
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

func (storage *FileStorage) DeleteAll() error {
	for fn := range storage.files {
		err := storage.files[fn].Delete()
		if err != nil {
			return err
		}
	}

	storage.files = make(map[string]*File)

	return nil
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

	_, err := file.f.Seek(offset, 0)
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
