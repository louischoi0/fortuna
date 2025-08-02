package model

import (
	"errors"
	"fmt"
	"strconv"
)

const OP_SYMBOL = '*'
const ARG_SYMBOL = '$'
const STR_SYMBOL = '^'
const INT_SYMBOL = '!'

// ParseCompactOperation parses compact symbolic string (e.g., "*&1$^abc") into OperationRaw
func ParseCompactOperation(input string) (*OperationRaw, error) {
	p := &reader{src: input, pos: 0}
	return parseOperation(p)
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

	if r.next() != ARG_SYMBOL {
		return nil, fmt.Errorf("expected '&' after '*', got %q", r.peek())
	}

	// parse OpCode number
	code := r.readWhile(func(c byte) bool { return c >= '0' && c <= '9' })
	if code == "" {
		return nil, errors.New("missing op code number after '&'")
	}
	op := &OperationRaw{
		OpCode: string(ARG_SYMBOL) + code,
		Args:   []interface{}{},
	}

	// parse args
	for {
		ch := r.peek()
		if ch == 0 || ch == OP_SYMBOL {
			break // next operation or end
		}

		if r.next() != ARG_SYMBOL {
			return nil, fmt.Errorf("expected '$' before argument, got %q", ch)
		}

		switch r.peek() {
		case STR_SYMBOL: // string
			r.next()
			str := r.readWhile(func(c byte) bool { return c != ARG_SYMBOL && c != OP_SYMBOL })
			if str == "" {
				return nil, fmt.Errorf("empty string value after '^'")
			}
			op.Args = append(op.Args, str)

		case INT_SYMBOL: // integer
			r.next()
			numStr := r.readWhile(func(c byte) bool { return c >= '0' && c <= '9' })
			if numStr == "" {
				return nil, fmt.Errorf("empty int value after '!'")
			}
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer: %v", err)
			}
			op.Args = append(op.Args, num)

		case OP_SYMBOL: // nested operation
			nested, err := parseOperation(r)
			if err != nil {
				return nil, fmt.Errorf("nested operation parse failed: %v", err)
			}
			op.Args = append(op.Args, nested)

		default:
			return nil, fmt.Errorf("unexpected character after $: %q", r.peek())
		}
	}

	return op, nil
}
