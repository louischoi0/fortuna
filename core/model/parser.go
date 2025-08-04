package model

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	OP_SYMBOL     = '*'
	OP_END_SYMBOL = ';'

	CODE_MARK = 'x'
	ARG_MARK  = '$'

	STR_SYMBOL     = '"'
	STR_END_SYMBOL = '"'
	INT_SYMBOL     = '!'
	INT_END_SYMBOL = '!'
	VEC_SYMBOL     = '%'
	VEC_END_SYMBOL = '%'
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
		switch ch {
		case 0:
			return nil, errors.New("unterminated operation: missing ';'")
		case OP_END_SYMBOL:
			r.next() // consume ';'
			return op, nil
		default:
			// expect an argument
			if r.next() != ARG_MARK {
				return nil, fmt.Errorf("expected '$' before argument, got %q", ch)
			}

			switch r.peek() {
			case STR_SYMBOL:
				// $"..."`
				if r.next() != STR_SYMBOL {
					return nil, errors.New(`expected '"' to start string`)
				}
				str := r.readWhile(func(c byte) bool {
					return c != STR_END_SYMBOL
				})
				if r.next() != STR_END_SYMBOL {
					return nil, errors.New(`unterminated string: missing '"'`)
				}
				op.Args = append(op.Args, str)

			case INT_SYMBOL:
				// $!123!
				if r.next() != INT_SYMBOL {
					return nil, errors.New("expected '!' to start int")
				}
				numStr := r.readWhile(func(c byte) bool { return c >= '0' && c <= '9' })
				if numStr == "" {
					return nil, fmt.Errorf("empty int after '!'")
				}
				if r.next() != INT_END_SYMBOL {
					return nil, fmt.Errorf("unterminated int: expected closing '%c'", INT_END_SYMBOL)
				}
				num, err := strconv.ParseInt(numStr, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid integer: %v", err)
				}
				op.Args = append(op.Args, num)

			case OP_SYMBOL:
				// nested $*xNN ... ;
				nested, err := parseOperation(r)
				if err != nil {
					return nil, fmt.Errorf("failed to parse nested operation: %w", err)
				}
				op.Args = append(op.Args, nested)

			case VEC_SYMBOL:
				// TODO: VECTOR 파싱은 추후 구현
				return nil, errors.New("TODO: vector parsing not implemented yet")

			default:
				return nil, fmt.Errorf("unexpected character after $: %q", r.peek())
			}
		}
	}
}

func ParseCompactOperations(input string) ([]*OperationRaw, error) {
	r := &reader{src: input, pos: 0}
	var ops []*OperationRaw

	for {
		// 공백 스킵
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
