package model

import (
	"fortuna/crypto"
	"fortuna/structure"
	"strconv"
	"strings"
	"time"
)

type Transaction struct {
	SpaceID    string                `json:"space_id"`
	From       string                `json:"from"`
	Params     *structure.OrderedMap `json:"params"`
	Operations []*Operation          `json:"operations"`
	Timestamp  int64                 `json:"timestamp"`
	RawCode    []byte                `json:"raw_code"`
}

func NewTransaction(spaceID string, from string, params *structure.OrderedMap) *Transaction {
	return &Transaction{
		Timestamp:  time.Now().Unix(),
		Operations: []*Operation{},
		RawCode:    []byte{},
	}
}

func NewLogMachineStateTransaction(spaceID string, from string, params *structure.OrderedMap) *Transaction {
	return &Transaction{
		SpaceID: spaceID,
		From:    from,
		Params:  params,
	}
}

func (t *Transaction) Hash() string {
	var buf strings.Builder

	t.UpdateRawCode()

	buf.WriteString(strconv.Itoa(int(t.Timestamp)))
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(t.SpaceID)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(t.From)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(t.Params.Hash())
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(string(t.RawCode))

	return crypto.SHA256(buf.String())
}

func (t *Transaction) AddOperation(op *Operation) {
	t.Operations = append(t.Operations, op)
}

func (t *Transaction) UpdateRawCode() ([]byte, error) {
	compact_ops, err := SerializeCompactOperations(t.Operations)
	if err != nil {
		return nil, err
	}

	t.RawCode = []byte(compact_ops)
	return t.RawCode, nil
}
