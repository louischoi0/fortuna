package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	OP_SYMBOL  = '*'
	CODE_MARK  = 'x'
	ARG_SYMBOL = '$'
	STR_SYMBOL = '^'
	INT_SYMBOL = '!'
)

// ParseCompactOperation parses a compact symbolic string into an OperationRaw
func ParseCompactOperation(input string) (*OperationRaw, error) {
	r := &reader{src: input, pos: 0}
	return parseOperation(r)
}

type reader struct {
	src string
	pos int
}

func (r *reader) peek() byte {
	if r.pos >= len(r.src) {
		return 0
	}
	return r.src[r.pos]
}

func (r *reader) next() byte {
	if r.pos >= len(r.src) {
		return 0
	}
	ch := r.src[r.pos]
	r.pos++
	return ch
}

func (r *reader) readWhile(pred func(byte) bool) string {
	start := r.pos
	for r.pos < len(r.src) && pred(r.src[r.pos]) {
		r.pos++
	}
	return r.src[start:r.pos]
}

func parseOperation(r *reader) (*OperationRaw, error) {
	if r.next() != OP_SYMBOL {
		return nil, fmt.Errorf("expected '*', got %q", r.peek())
	}

	if r.next() != CODE_MARK {
		return nil, fmt.Errorf("expected 'x' after '*', got %q", r.peek())
	}

	// read opcode number (e.g., x00 → "x00")
	code := r.readWhile(func(c byte) bool { return c >= '0' && c <= '9' })
	if code == "" {
		return nil, errors.New("missing opcode number after 'x'")
	}

	op := &OperationRaw{
		OpCode: string(CODE_MARK) + code,
		Args:   []interface{}{},
	}

	for {
		ch := r.peek()
		if ch == 0 || ch == OP_SYMBOL {
			break
		}

		if r.next() != ARG_SYMBOL {
			return nil, fmt.Errorf("expected '$' before argument, got %q", ch)
		}

		switch r.peek() {
		case STR_SYMBOL:
			r.next()
			str := r.readWhile(func(c byte) bool {
				return c != ARG_SYMBOL && c != OP_SYMBOL
			})
			if str == "" {
				return nil, fmt.Errorf("empty string after '^'")
			}
			op.Args = append(op.Args, str)

		case INT_SYMBOL:
			r.next()
			numStr := r.readWhile(func(c byte) bool { return c >= '0' && c <= '9' })
			if numStr == "" {
				return nil, fmt.Errorf("empty int after '!'")
			}
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer: %v", err)
			}
			op.Args = append(op.Args, num)

		case OP_SYMBOL:
			nested, err := parseOperation(r)
			if err != nil {
				return nil, fmt.Errorf("failed to parse nested operation: %w", err)
			}
			op.Args = append(op.Args, nested)

		default:
			return nil, fmt.Errorf("unexpected character after $: %q", r.peek())
		}
	}

	return op, nil
}

func ParseCompactOperations(input string) ([]*OperationRaw, error) {
	r := &reader{src: input, pos: 0}
	var ops []*OperationRaw

	for {
		for r.peek() == ' ' || r.peek() == '\n' || r.peek() == '\t' || r.peek() == '\r' {
			r.next()
		}

		if r.pos >= len(r.src) {
			break
		}

		op, err := parseOperation(r)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func SerializeCompactOperations(ops []*OperationRaw) (string, error) {
	var sb strings.Builder
	for _, op := range ops {
		if err := serializeOperation(&sb, op); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

func serializeOperation(sb *strings.Builder, op *OperationRaw) error {
	if op == nil {
		return fmt.Errorf("nil operation")
	}

	// start flag
	sb.WriteByte(OP_SYMBOL)

	// OpCode
	sb.WriteString(op.OpCode)

	// Args
	for _, arg := range op.Args {
		sb.WriteByte(ARG_SYMBOL) // '$'
		switch v := arg.(type) {
		case string:
			sb.WriteByte(STR_SYMBOL) // '^'
			sb.WriteString(v)
		case int64:
			sb.WriteByte(INT_SYMBOL) // '!'
			sb.WriteString(strconv.FormatInt(v, 10))
		case *OperationRaw:
			if err := serializeOperation(sb, v); err != nil {
				return fmt.Errorf("nested operation serialize failed: %v", err)
			}
		default:
			return fmt.Errorf("unsupported argument type: %T", v)
		}
	}
	return nil
}
