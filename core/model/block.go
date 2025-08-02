package model

import (
	"fortuna/crypto"
	"fortuna/structure"
	"strings"
	"sync"
)

type EventBlock struct {
	mu sync.Mutex

	SpaceID           string
	Height            int64
	Count             int64
	PreviousBlockHash string
	PreviousBlock     *EventBlock

	ExecutionRootHash   string
	TransactionRootHash string
	UniverseHash        string

	Executions   []*EventExecutionResult
	Transactions []interface{}
}

func (block *EventBlock) AppendTransactionExecution(tx interface{}) {
	block.mu.Lock()
	defer block.mu.Unlock()

	block.Transactions = append(block.Transactions, tx)
}

func (block *EventBlock) AppendEventExecution(er *EventExecutionResult) {
	block.mu.Lock()
	defer block.mu.Unlock()

	block.Executions = append(block.Executions, er)
}

func (block *EventBlock) Hash() string {
	var buf strings.Builder

	buf.WriteString(block.PreviousBlockHash)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(block.ExecutionRootHash)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(block.TransactionRootHash)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(block.UniverseHash)

	return crypto.SHA256(buf.String())
}

func (block *EventBlock) Update() {
	contents := make([]structure.Content, len(block.Executions))
	for i, er := range block.Executions {
		contents[i] = er
	}
	mt, err := structure.NewTree(contents)
	if err != nil {
		panic(err)
	}
	block.ExecutionRootHash = string(mt.MerkleRoot())
}
