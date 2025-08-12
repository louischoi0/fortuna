package test

import (
	"fortuna/core/storage"
	"testing"
	"fmt"
)

func TestFileStorage(t *testing.T) {
	engine := storage.NewFileStorage("abcdefg")
	file := storage.NewFile(engine, "a")

	buffer := []byte("abcdefgeabcdefgeabcdefgeabcdefgeabcdefgeabcdefgeabcdefgeabcdefgeabcdefgeabcdefgeabcdefgeabcdefge")

	file.AppendFileBytes(buffer)
	ret, err := file.OffsetRead(0, 1)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("ret: %s", string(ret))
}
