package model

import (
	"fortuna/crypto"
	"fortuna/structure"
	"fortuna/util"
	"strings"
	"fmt"
	"bytes"
	"sync"
)

const BLOCK_HASH_STR_LENGTH = 64

type Block struct {
	mu 			sync.Mutex

	SpaceID           	string
	Timestamp		int64

	Height            	int64
	Count             	int64
	PreviousBlockHash 	string
	PreviousBlock     	*Block

	ExecutionRootHash   	string
	TransactionRootHash 	string
	UniverseHash        	string

	Executions   		[]*EventExecutionResult
	Transactions 		[]*Transaction
}

func NewBlock(spaceID string, height int64, prevBlock *Block) *Block {
	var phash string
	if prevBlock != nil {
		phash = prevBlock.Hash()
	}

	return &Block{
		SpaceID: spaceID,
		Height: height, 
		Count: 0,
		PreviousBlock: prevBlock,
		PreviousBlockHash: phash,
		Executions: make([]*EventExecutionResult, 0, 50),
		Transactions: make([]*Transaction, 0, 30),
	}
}

func (block *Block) AppendTransactionExecution(tx *Transaction) {
	block.mu.Lock()
	defer block.mu.Unlock()

	block.Transactions = append(block.Transactions, tx)
}

func (block *Block) AppendEventExecution(er *EventExecutionResult) {
	block.mu.Lock()
	defer block.mu.Unlock()

	block.Executions = append(block.Executions, er)
}

func (block *Block) Hash() string {
	var buf strings.Builder

	buf.Write(util.EncodeUint64(uint64(block.Height)))
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(block.PreviousBlockHash)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(block.ExecutionRootHash)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(block.TransactionRootHash)
	buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(block.UniverseHash)

	return crypto.SHA256(buf.String())
}

func (block *Block) UpdateTransactionRoot() {
	contents := make([]structure.Content, len(block.Transactions))
	for i, er := range block.Transactions {
		contents[i] = er
	}
	mt, err := structure.NewTree(contents)
	if err != nil {
		panic(err)
	}
	block.TransactionRootHash = string(mt.MerkleRoot())
}

func (block *Block) UpdateExecutionRoot() {
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

func (block *Block) Verify() error {
	if !(len(block.Transactions) == 0 && len(block.Executions) == 0) {
		return fmt.Errorf("block has no transaction and event")
	}

	if block.Height == 0 {
		return fmt.Errorf("block has height zero")
	}

	if block.Height != 1 && block.PreviousBlockHash == "" {
		return fmt.Errorf("block has height zero")
	}
	return nil
}

func (block *Block) Encode() ([]byte, error) {
	if err := block.Verify(); err != nil {
		return nil, err
	}

	hash := block.Hash()	
	var buf bytes.Buffer

	if len(hash) != BLOCK_HASH_STR_LENGTH {
		return nil, fmt.Errorf("block has invalid hash: %v", hash)
	}

	buf.WriteString(hash)
	buf.Write(util.EncodeUint64(uint64(block.Timestamp)))
	
	if len(block.PreviousBlockHash) != BLOCK_HASH_STR_LENGTH  {
		return nil, fmt.Errorf("block has invalid hash: %v", hash)
	}

	buf.WriteString(block.PreviousBlockHash)

	txCount := len(block.Transactions)
	EventCount := len(block.Executions)

	// Count Flag
	buf.Write(util.EncodeUint64(uint64(EventCount)))
	buf.Write(util.EncodeUint64(uint64(txCount)))
	
	for _, event := range(block.Executions) {
		eventBuffer, err := event.Encode()
		eventBufferSize := len(eventBuffer)
		if err != nil {
			return nil, err
		}
		buf.Write(util.EncodeUint64(uint64(eventBufferSize)))
		buf.Write(eventBuffer)
	}

	for _, tx := range(block.Transactions) {
		txBuffer, err := tx.Encode()
		txBufferSize := len(txBuffer)
		if err != nil {
			return nil, err
		}
		buf.Write(util.EncodeUint64(uint64(txBufferSize)))
		buf.Write(txBuffer)
	}
	return buf.Bytes(), nil
}

