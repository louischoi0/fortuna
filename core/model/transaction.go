package model

import (
	"fortuna/crypto"
	"fortuna/structure"
	"fortuna/util"
	"strconv"
	"strings"
	"bytes"
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
		Timestamp:  util.Now(),
		Operations: []*Operation{},
		RawCode:    []byte{},
	}
}

func (t *Transaction) Encode() []byte {
	var buf bytes.Buffer
	return buf.Bytes()
}

func (tx *Transaction) Verify(hash string) bool {
	return tx.Hash() ==  hash 
}

func (tx *Transaction) Hash() string {
	var buf strings.Builder

	tx.UpdateRawCode()

	buf.WriteString(strconv.Itoa(int(tx.Timestamp)))
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(tx.SpaceID)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(tx.From)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(tx.Params.Hash())
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(string(tx.RawCode))

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
