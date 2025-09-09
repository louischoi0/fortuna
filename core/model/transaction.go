package model

import (
	"bytes"
	"fmt"
	"fortuna/crypto"
	"fortuna/util"
	"strconv"
	"strings"

	C "fortuna/core/config"
)

type Transaction = RawTransaction

type RawTransaction struct {
	Type	   string	`json:"type"`
	SpaceID    string       `json:"space_id"`
	From       string       `json:"from"`
	Operations []*Operation `json:"operations"`
	Timestamp  int64        `json:"timestamp"`
	RawCode    []byte       `json:"raw_code"`
}

func NewRawTransaction(spaceID string, from string) *RawTransaction {
	return &RawTransaction{
		SpaceID:    spaceID,
		From:       from,
		Timestamp:  util.Now(),
		Operations: []*Operation{},
		RawCode:    []byte{},
	}
}
func (tx *RawTransaction) SetType(txType string) {
	tx.Type = txType
}

func (tx *RawTransaction) Verify(hash string) error {
	if len(hash) != C.MODEL_HASH_STR_LENGTH {
		return fmt.Errorf("hash must have length %d", C.MODEL_HASH_STR_LENGTH)
	}

	if len(tx.SpaceID) != C.SPACE_ID_STR_LENGTH {
		return fmt.Errorf("spaceID must have length %d", C.SPACE_ID_STR_LENGTH)
	}

	if len(tx.From) != C.IDENTITY_ADDRESS_STR_LENGTH {
		return fmt.Errorf("from must have length %d", C.IDENTITY_ADDRESS_STR_LENGTH)
	}

	if len(tx.RawCode) == 0 {
		return fmt.Errorf("rawCode must have length %d", len(tx.RawCode))
	}

	if len(tx.Operations) == 0 {
		return fmt.Errorf("operations must have length %d", len(tx.Operations))
	}

	if tx.Hash() != hash {
		return fmt.Errorf("hash mismatch")
	}

	return nil
}

func (tx *RawTransaction) Hash() string {
	var buf strings.Builder
	tx.UpdateRawCode()

	buf.WriteString(strconv.Itoa(int(tx.Timestamp)))
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(tx.SpaceID)
	buf.WriteString(C.HASH_SEPERATOR)
	buf.WriteString(tx.From)
	buf.WriteString(C.HASH_SEPERATOR)
	// buf.WriteString(tx.Params.Hash())
	// buf.WriteString(HASH_SEPERATOR)
	buf.WriteString(string(tx.RawCode))

	return crypto.SHA256(buf.String())
}

func (t *RawTransaction) AddOperation(op *Operation) {
	t.Operations = append(t.Operations, op)
}

func (t *RawTransaction) UpdateRawCode() ([]byte, error) {
	compact_ops, err := SerializeCompactOperations(t.Operations)
	if err != nil {
		return nil, err
	}

	t.RawCode = []byte(compact_ops)
	return t.RawCode, nil
}

func (rtx *RawTransaction) Encode() ([]byte, error) {
	var buf bytes.Buffer
	rtx.UpdateRawCode()

	hash := rtx.Hash()
	if len(hash) != C.MODEL_HASH_STR_LENGTH {
		return nil, fmt.Errorf("hash must have length %d", C.MODEL_HASH_STR_LENGTH)
	}

	if len(rtx.SpaceID) != C.SPACE_ID_STR_LENGTH {
		return nil, fmt.Errorf("spaceID must have length %d", C.SPACE_ID_STR_LENGTH)
	}

	if len(rtx.From) != C.IDENTITY_ADDRESS_STR_LENGTH {
		return nil, fmt.Errorf("from must have length %d", C.IDENTITY_ADDRESS_STR_LENGTH)
	}

	buf.WriteString(hash)
	buf.Write(util.EncodeUint64(uint64(rtx.Timestamp)))

	buf.WriteString(rtx.SpaceID)
	buf.WriteString(rtx.From)

	buf.WriteByte(1)
	rawCodeBufferSize := uint64(len(rtx.RawCode))
	buf.Write(util.EncodeUint64(rawCodeBufferSize))

	buf.Write(rtx.RawCode)

	return buf.Bytes(), nil
}

func DecodeRawTransaction(b []byte) (*RawTransaction, error) {
	var (
		off = 0
		n   = len(b)
	)

	need := func(k int) error {
		if off+k > n {
			return fmt.Errorf("buffer underflow: %d", k)
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
		u, err := util.DecodeUint64(b[off : off+8])
		if err != nil {
			return 0, err
		}
		off += 8
		return u, nil
	}

	_, err := readFixedString(C.MODEL_HASH_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read hash: %w", err)
	}

	ts, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read timestamp: %w", err)
	}

	spaceID, err := readFixedString(C.SPACE_ID_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read spaceID: %w", err)
	}

	from, err := readFixedString(C.IDENTITY_ADDRESS_STR_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("read from: %w", err)
	}


	if err := need(1); err != nil {
		return nil, fmt.Errorf("read raw_code flag: %w", err)
	}
	rawCodeFlag := b[off]
	off++

	if rawCodeFlag != 1 {
		return nil, fmt.Errorf("unexpected raw_code flag: %d at %d:%d", rawCodeFlag, off-1, len(b))
	}
	rawCodeLen, err := readU64LE()
	if err != nil {
		return nil, fmt.Errorf("read raw_code length: %w", err)
	}
	if err := need(int(rawCodeLen)); err != nil {
		return nil, fmt.Errorf("read raw_code body: %w", err)
	}
	rawCode := b[off : off+int(rawCodeLen)]
	off += int(rawCodeLen)

	ops, err := ParseCompactOperations(rawCode)
	if err != nil {
		return nil, fmt.Errorf("parse ops: %w", err)
	}

	return &RawTransaction{
		Timestamp:  int64(ts),
		SpaceID:    spaceID,
		From:       from,
		RawCode:    rawCode,
		Operations: ops,
	}, nil
}
