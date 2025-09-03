package model

import (
	"bytes"
	"fmt"
	"fortuna/crypto"
	"fortuna/structure"
	"fortuna/util"
	"log"
	"strings"
	"sync"
	"encoding/json"

	C "fortuna/core/config"
)

type Page struct {
	mu sync.Mutex

	SpaceID   string
	Timestamp uint64

	N            uint64
	PrevPageHash string
	PrevPage     *Page
	Committed	bool

	ExecutionRootHash   string
	TransactionRootHash string
	UniverseHash        string

	Executions   []*EventResult
	Transactions []*Transaction
}

func NewPage(spaceID string, pageNum uint64, prevPage *Page) *Page {
	var phash string
	if prevPage != nil {
		phash = prevPage.Hash()
	} else {
		phash = C.ZERO_HASH
	}

	return &Page{
		SpaceID:             spaceID,
		N:                   pageNum,
		PrevPage:            prevPage,
		PrevPageHash:        phash,
		Executions:          make([]*EventResult, 0, 50),
		Transactions:        make([]*Transaction, 0, 25),
		TransactionRootHash: C.ZERO_HASH,
		ExecutionRootHash:   C.ZERO_HASH,
		UniverseHash:        C.ZERO_HASH,
		Committed: 	     false,
	}
}

func (page *Page) GetCount() int {
	return len(page.Executions) + len(page.Transactions)
}

func (page *Page) AppendTransaction(tx *Transaction) {
	page.mu.Lock()
	defer page.mu.Unlock()

	log.Printf("append raw transaction for page %v: %v", page.N, tx.Hash())
	page.Transactions = append(page.Transactions, tx)
}

func (page *Page) AppendEventExecution(er *EventResult) {
	page.mu.Lock()
	defer page.mu.Unlock()
	log.Printf("append event execution for page %v: %v", page.N, er.Hash())

	page.Executions = append(page.Executions, er)
}

func (page *Page) Hash() string {
	var buf strings.Builder

	buf.Write(util.EncodeUint64(uint64(page.N)))
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(page.PrevPageHash)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(page.ExecutionRootHash)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(page.TransactionRootHash)

	return crypto.SHA256(buf.String())
}

func (page *Page) UpdateTransactionRoot() {
	if page.Committed {
		log.Fatalf("updating hash committed page is not allowed")
	}

	if len(page.Transactions) == 0 {
		page.TransactionRootHash = C.ZERO_HASH
		return
	}

	contents := make([]structure.Content, len(page.Transactions))
	for i, er := range page.Transactions {
		contents[i] = er
	}
	mt, err := structure.NewTree(contents)
	if err != nil {
		panic(err)
	}
	page.TransactionRootHash = string(mt.MerkleRoot())
}

func (page *Page) UpdateExecutionRoot() {
	if page.Committed {
		log.Fatalf("updating hash committed page is not allowed")
	}

	if len(page.Executions) == 0 {
		page.ExecutionRootHash = C.ZERO_HASH
		return
	}

	contents := make([]structure.Content, len(page.Executions))
	for i, er := range page.Executions {
		contents[i] = er
	}
	mt, err := structure.NewTree(contents)
	if err != nil {
		panic(err)
	}
	page.ExecutionRootHash = string(mt.MerkleRoot())
}

func (page *Page) Verify() error {
	if len(page.Transactions) == 0 && len(page.Executions) == 0 {
		return fmt.Errorf("page has no transaction and event")
	}

	if page.N == 0 {
		return fmt.Errorf("page has num zero")
	}

	if page.N != 1 && page.PrevPageHash == "" {
		return fmt.Errorf("page has num zero")
	}
	return nil
}

func (page *Page) Encode() ([]byte, error) {
	if err := page.Verify(); err != nil {
		return nil, err
	}

	hash := page.Hash()

	if len(hash) != C.MODEL_HASH_STR_LENGTH {
		return nil, fmt.Errorf("page has invalid hash: %v", hash)
	}

	if page.N > 1 && len(page.PrevPageHash) != C.MODEL_HASH_STR_LENGTH {
		return nil, fmt.Errorf("page has invalid previous page hash: %v", hash)
	}

	var buf bytes.Buffer
	buf.WriteString(hash)
	buf.WriteString(page.SpaceID)
	buf.Write(util.EncodeUint64(uint64(page.N)))
	buf.Write(util.EncodeUint64(uint64(page.Timestamp)))

	buf.WriteString(page.PrevPageHash)
	buf.WriteString(page.TransactionRootHash)
	buf.WriteString(page.ExecutionRootHash)

	txCount := len(page.Transactions)
	EventCount := len(page.Executions)

	buf.Write(util.EncodeUint64(uint64(EventCount)))
	buf.Write(util.EncodeUint64(uint64(txCount)))

	for _, event := range page.Executions {
		eventBuffer, err := event.Encode()
		eventBufferSize := len(eventBuffer)
		if err != nil {
			return nil, err
		}
		buf.Write(util.EncodeUint64(uint64(eventBufferSize)))
		buf.Write(eventBuffer)
	}

	for _, tx := range page.Transactions {
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

func DecodePage(b []byte) (*Page, error) {
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

	// 1. page hash (skip validation)
	_, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, err
	}

	spaceID, err := readFixedString(64) //TODO
	if err != nil {
		return nil, err
	}

	pageNum, err := readU64LE()
	if err != nil {
		return nil, err
	}

	// 2. timestamp
	ts, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read timestamp: %w", err)
	}

	// 3. previous page hash
	prevHash, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read previous hash: %w", err)
	}

	// 4. root hashes
	txRootHash, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read transaction root hash: %w", err)
	}

	execRootHash, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read execution root hash: %w", err)
	}

	// 5. event count and tx count
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
		log.Printf("read event from page size: %d, offset: %d, len: %d", int(sizeU), off, len(data))

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
		log.Printf("read transaction from page size: %d, offset: %d, len: %d", int(sizeU), off, len(b))
		off += int(sizeU)

		tx, err := DecodeRawTransaction(data)
		if err != nil {
			return nil, fmt.Errorf("decode transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	page := &Page{
		SpaceID:	     spaceID,
		N:                   pageNum,
		Timestamp:           ts,
		PrevPageHash:        prevHash,
		TransactionRootHash: txRootHash,
		ExecutionRootHash:   execRootHash,
		Executions:          events,
		Transactions:        txs,
	}

	page.UpdateExecutionRoot()
	page.UpdateTransactionRoot()

	return page, nil
}

func (page *Page) Update() {
	if page.Committed {
		log.Fatalf("updating hash committed page is not allowed")
	}
	page.UpdateExecutionRoot()
	page.UpdateTransactionRoot()
}

func (page *Page) GetPageNum() uint64 {
	return page.N
}

func (page *Page) Show() string {
	buf, err := json.Marshal(page)
	if err != nil {
		return ""
	}

	log.Println("----------------------------------------")
	log.Printf(string(buf))
	log.Println("----------------------------------------")
	return string(buf)
}
