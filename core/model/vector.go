package model

import (
	"bytes"
	"fmt"
	"fortuna/util"
)

type StateVector struct {
	Count int
	Data  []int64
}

func NewStateVector(count int, data []int64) *StateVector {
	return &StateVector{
		Count: count,
		Data:  data,
	}
}

func (v *StateVector) Get(index int) int64 {
	return v.Data[index]
}

func (v *StateVector) Set(index int, value int64) {
	v.Data[index] = value
}

func (v *StateVector) Encode() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(VEC_SYMBOL)

	buffer := util.EncodeInt64Array(v.Data)
	buffer_size := len(buffer)
	buf.Write(util.EncodeUint32(uint32(buffer_size)))
	buf.Write(buffer)
	buf.WriteByte(VEC_END_SYMBOL)

	return buf.Bytes()
}

func DecodeStateVector(b []byte) (*StateVector, error) {
	if len(b) < 1+4+1 {
		return nil, fmt.Errorf("invalid state vector: too short")
	}
	i := 0

	if b[i] != VEC_SYMBOL {
		return nil, fmt.Errorf("invalid state vector: missing start symbol '%c'", VEC_SYMBOL)
	}
	i++

	if len(b) < i+4 {
		return nil, fmt.Errorf("invalid state vector: missing length (uint32)")
	}
	size, err := util.DecodeUint32(b[i : i+4])
	if err != nil {
		return nil, fmt.Errorf("invalid state vector: missing length (uint32)")
	}
	i += 4

	if size == 0 {
		if len(b) < i+1 {
			return nil, fmt.Errorf("invalid state vector: missing end symbol")
		}
		if b[i] != VEC_END_SYMBOL {
			return nil, fmt.Errorf("invalid state vector: expected end symbol '%c'", VEC_END_SYMBOL)
		}
		return &StateVector{Count: 0, Data: []int64{}}, nil
	}

	if len(b) < i+int(size) {
		return nil, fmt.Errorf("invalid state vector: payload truncated (need %d, have %d)", size, len(b)-i)
	}
	payload := b[i : i+int(size)]
	i += int(size)

	if len(b) < i+1 {
		return nil, fmt.Errorf("invalid state vector: missing end symbol")
	}
	if b[i] != VEC_END_SYMBOL {
		return nil, fmt.Errorf("invalid state vector: expected end symbol '%c'", VEC_END_SYMBOL)
	}
	i++

	// 5) decode int64 array from payload
	data, err := util.DecodeInt64Array(payload)
	if err != nil {
		return nil, fmt.Errorf("invalid state vector: failed to decode int64 array: %w", err)
	}

	return &StateVector{
		Count: len(data),
		Data:  data,
	}, nil
}
