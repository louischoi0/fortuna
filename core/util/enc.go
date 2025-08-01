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

func EncodeInt64(n int64) []byte {
    buf := make([]byte, 8)
    binary.BigEndian.PutUint64(buf, uint64(n))
    return buf
}

func DecodeInt64(b []byte) (int64, error) {
    if len(b) != 8 {
        return 0, errors.New("bytes must has length 8")
    }
    num := int64(binary.BigEndian.Uint64(b))
    return num, nil
}

func EncodeInt32(value int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeInt32(data []byte) (int32, error) {
	var value int32
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func EncodeUint32(n uint32) []byte {
    buf := make([]byte, 4)
    binary.BigEndian.PutUint32(buf, n)
    return buf
}

func DecodeUint32(b []byte) (uint32, error) {
    if len(b) != 4 {
        return 0, fmt.Errorf("invalid byte length for uint32: expected 4, got %d", len(b))
    }
    return binary.BigEndian.Uint32(b), nil
}
