package util

import (
	"bytes"
	"encoding/binary"
)

func EncodeInt64Array(arr []int64) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, arr)
	return buf.Bytes()
}

func DecodeInt64Array(data []byte) []int64 {
	count := len(data)
	result := make([]int64, count)
	binary.Read(bytes.NewReader(data), binary.LittleEndian, &result)
	return result
}
