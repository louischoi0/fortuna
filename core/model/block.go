package model

import (
	"bytes"
	"fmt"
	"fortuna/crypto"
	"fortuna/structure"
	"fortuna/util"
	"strings"
	"sync"

	C "fortuna/core/config"
)

type Block struct {
	mu sync.Mutex

	SpaceID   string
	Timestamp int64

	Height            int64
	Count             int64
	PreviousBlockHash string
	PreviousBlock     *Block

	ExecutionRootHash   string
	TransactionRootHash string
	UniverseHash        string

	Executions   []*EventResult
	Transactions []*Transaction
}

func NewBlock(spaceID string, height int64, prevBlock *Block) *Block {
	var phash string
	if prevBlock != nil {
		phash = prevBlock.Hash()
	} else {
		phash = C.ZERO_HASH
	}

	return &Block{
		SpaceID:             spaceID,
		Height:              height,
		Count:               0,
		PreviousBlock:       prevBlock,
		PreviousBlockHash:   phash,
		Executions:          make([]*EventResult, 0, 50),
		Transactions:        make([]*Transaction, 0, 25),
		TransactionRootHash: C.ZERO_HASH,
		ExecutionRootHash:   C.ZERO_HASH,
		UniverseHash:        C.ZERO_HASH,
	}
}

func (block *Block) AppendTransactionExecution(tx *Transaction) {
	block.mu.Lock()
	defer block.mu.Unlock()

	block.Transactions = append(block.Transactions, tx)
}

func (block *Block) AppendEventExecution(er *EventResult) {
	block.mu.Lock()
	defer block.mu.Unlock()

	block.Executions = append(block.Executions, er)
}

func (block *Block) Hash() string {
	var buf strings.Builder

	buf.Write(util.EncodeUint64(uint64(block.Height)))
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(block.PreviousBlockHash)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(block.ExecutionRootHash)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(block.TransactionRootHash)
	// buf.WriteString(HASH_SEPERATOR)
	// buf.WriteString(block.UniverseHash)

	return crypto.SHA256(buf.String())
}

func (block *Block) UpdateTransactionRoot() {
	if len(block.Transactions) == 0 {
		block.TransactionRootHash = C.ZERO_HASH
		return
	}

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
	if len(block.Executions) == 0 {
		block.ExecutionRootHash = C.ZERO_HASH
		return
	}

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
	if len(block.Transactions) == 0 && len(block.Executions) == 0 {
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

	if len(hash) != C.MODEL_HASH_STR_LENGTH {
		return nil, fmt.Errorf("block has invalid hash: %v", hash)
	}

	if block.Height > 1 && len(block.PreviousBlockHash) != C.MODEL_HASH_STR_LENGTH {
		return nil, fmt.Errorf("block has invalid previous block hash: %v", hash)
	}

	var buf bytes.Buffer
	buf.WriteString(hash)
	buf.Write(util.EncodeUint64(uint64(block.Height)))
	buf.Write(util.EncodeUint64(uint64(block.Timestamp)))

	buf.WriteString(block.PreviousBlockHash)
	// buf.WriteString(block.TransactionRootHash)
	// buf.WriteString(block.ExecutionRootHash)

	txCount := len(block.Transactions)
	EventCount := len(block.Executions)

	buf.Write(util.EncodeUint64(uint64(EventCount)))
	buf.Write(util.EncodeUint64(uint64(txCount)))

	for _, event := range block.Executions {
		eventBuffer, err := event.Encode()
		eventBufferSize := len(eventBuffer)
		if err != nil {
			return nil, err
		}
		buf.Write(util.EncodeUint64(uint64(eventBufferSize)))
		buf.Write(eventBuffer)
	}

	for _, tx := range block.Transactions {
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

func DecodeBlock(b []byte) (*Block, error) {
	var (
		off = 0
		n   = len(b)
	)

	need := func(k int) error {
		if off+k > n {
			return fmt.Errorf("buffer underflow: need %d bytes", k)
		}
		return nil
	}
	readFixedString := func(k int) (string, error) {
		if err := need(k); err != nil {
			return "", err
		}
		s := string(b[off : off+k])
		off += k
		return s, nil
	}
	readU64LE := func() (uint64, error) {
		if err := need(8); err != nil {
			return 0, err
		}
		v, err := util.DecodeUint64(b[off : off+8])
		if err != nil {
			return 0, err
		}
		off += 8
		return v, nil
	}

	// 1. block hash (skip validation)
	_, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read block hash: %w", err)
	}

	height, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read height: %w", err)
	}

	// 2. timestamp
	ts, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read timestamp: %w", err)
	}

	// 3. previous block hash
	prevHash, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read previous hash: %w", err)
	}

	// 4. event count and tx count
	eventCountU, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read event count: %w", err)
	}
	txCountU, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read tx count: %w", err)
	}

	eventCount := int(eventCountU)
	txCount := int(txCountU)

	// 5. decode event executions
	var events []*EventResult
	for i := 0; i < eventCount; i++ {
		sizeU, err := readU64LE()
		if err != nil {
			return nil, fmt.Errorf("read event size: %w", err)
		}
		if err := need(int(sizeU)); err != nil {
			return nil, fmt.Errorf("event buffer underflow: %w", err)
		}
		data := b[off : off+int(sizeU)]
		off += int(sizeU)

		evt, err := DecodeEventResult(data)
		if err != nil {
			return nil, fmt.Errorf("decode event: %w", err)
		}
		events = append(events, evt)
	}

	// 6. decode transactions
	var txs []*Transaction
	for i := 0; i < txCount; i++ {
		sizeU, err := readU64LE()
		if err != nil {
			return nil, fmt.Errorf("read tx size: %w", err)
		}
		if err := need(int(sizeU)); err != nil {
			return nil, fmt.Errorf("tx buffer underflow: %w", err)
		}
		data := b[off : off+int(sizeU)]
		off += int(sizeU)

		tx, err := DecodeRawTransaction(data)
		if err != nil {
			return nil, fmt.Errorf("decode transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	block := &Block{
		Height:            int64(height),
		Timestamp:         int64(ts),
		PreviousBlockHash: prevHash,
		Executions:        events,
		Transactions:      txs,
	}

	block.UpdateExecutionRoot()
	block.UpdateTransactionRoot()

	return block, nil
}
