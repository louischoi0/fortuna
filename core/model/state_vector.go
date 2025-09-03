package model

import (
	"bytes"
	"fmt"
	"fortuna/util"
)

type StateVector struct {
	Data []int64
}

func (v *StateVector) Var() *Var {
	return &Var{
		Value: v,
		Type: VarTypeVector,
	}
}

func NewStateVector(data []int64) *StateVector {
	return &StateVector{
		Data: data,
	}
}

func (v *StateVector) Get(index int64) int64 {
	return v.Data[index]
}

func (v *StateVector) Set(index int64, value int64) {
	v.Data[index] = value
}

func (v *StateVector) Encode() []byte {
	buf := bytes.NewBuffer(nil)

	buffer := util.EncodeInt64Array(v.Data)
	buffer_size := uint64(len(buffer))
	buf.Write(util.EncodeUint64(buffer_size))
	buf.Write(buffer)

	return buf.Bytes()
}

func DecodeStateVector(b []byte) (*StateVector, error) {
	if len(b) < 8 {
		return nil, fmt.Errorf("invalid state vector: too short")
	}
	i := 0

	if len(b) < i+8 {
		return nil, fmt.Errorf("invalid state vector: missing length (uint64)")
	}
	size, err := util.DecodeUint64(b[i : i+8])
	if err != nil {
		return nil, fmt.Errorf("invalid state vector: missing length (uint64)")
	}
	i += 8

	if size == 0 {
		return &StateVector{Data: []int64{}}, nil
	}

	if len(b) < i+int(size) {
		return nil, fmt.Errorf("invalid state vector: payload truncated (need %d, have %d)", size, len(b)-i)
	}
	payload := b[i : i+int(size)]
	i += int(size)

	data, err := util.DecodeInt64Array(payload)
	if err != nil {
		return nil, fmt.Errorf("invalid state vector: failed to decode int64 array: %w", err)
	}

	return &StateVector{
		Data: data,
	}, nil
}

func (v *StateVector) Size() int {
	return len(v.Data)
}
