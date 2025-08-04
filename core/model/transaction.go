package model

import (
	"fortuna/structure"
)

type OperationRaw struct {
	OpCode string        `json:"op_code"`
	OpName string        `json:"op_name"`
	Args   []interface{} `json:"args"`
}

type Transaction struct {
	SpaceID          string                `json:"space_id"`
	From             string                `json:"from"`
	Params           *structure.OrderedMap `json:"params"`
	Executed         bool
	ParsedOperations []*OperationRaw
	Timestamp	 int64
	Raw              []byte
}

func NewLogMachineStateTransaction(spaceID string, from string, params *structure.OrderedMap) *Transaction {
	return &Transaction{
		SpaceID: spaceID,
		From:    from,
		Params:  params,
	}
}
