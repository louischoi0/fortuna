package model

import (
	"fortuna/crypto"
	"strings"
	"sync"
)

type EventBlock struct {
	mu			sync.Mutex

	SpaceID			string
	Height			int64
	Count			int64
	PreviousBlockHash	string

	ExecutionRootHash	string
	TransactionRootHash	string
	UniverseHash		string

	Executions		[]*EventExecutionResult
	Transactions		[]interface{}
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
