package model

import (
	"errors"
	"fmt"
	"fortuna/util"
	"log"
	"strconv"
)

var (
	OP_SYMBOL     = []byte(`*`)
	OP_END_SYMBOL = []byte(`;`)

	CODE_MARK = []byte(`x`)
	ARG_MARK  = []byte(`$`)

	STR_SYMBOL     = []byte(`"`)
	STR_END_SYMBOL = []byte(`"`)
	INT_SYMBOL     = []byte(`!`)
	INT_END_SYMBOL = []byte(`!`)
	VEC_SYMBOL     = []byte(`%`)
	VEC_END_SYMBOL = []byte(`%`)
)

func asbyte(b []byte) byte {
	return b[0]
}

func ParseCompactOperation(input []byte) (*Operation, error) {
	r := &reader{src: input, pos: 0}
	return parseOperation(r)
}

type reader struct {
	src []byte
	pos int
}

func (r *reader) peek() byte {
	if r.pos >= len(r.src) {
		return 0
	}
	return r.src[r.pos]
}

func (r *reader) readU64LE() (uint64, error) {
	off := r.pos
	v, err := util.DecodeUint64(r.src[off : off+8])
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	r.pos += 8
	return v, nil
}

func (r *reader) take(n int) []byte {
	if r.pos+n > len(r.src) {
		return nil
	}

	ch := r.src[r.pos:r.pos+n]
	r.pos += n
	return ch
}

func (r *reader) next() byte {
	ch := r.src[r.pos]
	r.pos++
	return ch
}

func (r *reader) readWhile(pred func(byte) bool) []byte {
	start := r.pos
	for r.pos < len(r.src) && pred(r.src[r.pos]) {
		r.pos++
	}
	return r.src[start:r.pos]
}

func parseOperation(r *reader) (*Operation, error) {
	if r.next() != asbyte(OP_SYMBOL) {
		return nil, fmt.Errorf("expected '*', got %q", r.peek())
	}

	if r.next() != asbyte(CODE_MARK) {
		return nil, fmt.Errorf("expected 'x' after '*', got %q", r.peek())
	}

	code := r.readWhile(func(c byte) bool { return c >= '0' && c <= '9' })
	code_buf := make([]byte, 0, 100)
	code_buf = append(code_buf, CODE_MARK...)
	code_buf = append(code_buf, code...)

	op := &Operation{
		OpCode: string(code_buf),
		Args:   []interface{}{},
	}

	for {
		ch := r.peek()
		switch ch {
		case 0:
			return nil, errors.New("unterminated operation: missing ';'")
		case asbyte(OP_END_SYMBOL):
			r.next() // consume ';'
			return op, nil
		default:
			// expect an argument
			if r.next() != asbyte(ARG_MARK) {
				return nil, fmt.Errorf("expected '$' before argument, got %q", ch)
			}

			switch r.peek() {
			case asbyte(STR_SYMBOL):
				// $"..."`
				if r.next() != asbyte(STR_SYMBOL) {
					return nil, errors.New(`expected '"' to start string`)
				}
				str := r.readWhile(func(c byte) bool {
					return c != asbyte(STR_END_SYMBOL)
				})
				if r.next() != asbyte(STR_END_SYMBOL) {
					return nil, errors.New(`unterminated string: missing '"'`)
				}
				op.Args = append(op.Args, str)

			case asbyte(INT_SYMBOL):
				// $!123!
				if r.next() != asbyte(INT_SYMBOL) {
					return nil, errors.New("expected '!' to start int")
				}
				numStr := r.readWhile(func(c byte) bool { return c >= '0' && c <= '9' })
				if len(numStr) == 0 {
					return nil, fmt.Errorf("empty int after '!'")
				}
				if r.next() != asbyte(INT_END_SYMBOL) {
					return nil, fmt.Errorf("unterminated int: expected closing '%c'", INT_END_SYMBOL)
				}

				//TODO
				num, err := strconv.ParseInt(string(numStr), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid integer: %v", err)
				}
				op.Args = append(op.Args, num)

			case asbyte(OP_SYMBOL):
				// nested $*xNN ... ;
				nested, err := parseOperation(r)
				if err != nil {
					return nil, fmt.Errorf("failed to parse nested operation: %w", err)
				}
				op.Args = append(op.Args, nested)

			case asbyte(VEC_SYMBOL):
				if r.next() != asbyte(VEC_SYMBOL) {
					return nil, errors.New("expected '%' to start vec")
				}

				vec_buf_size, err := r.readU64LE()

				if err != nil {
					return nil, fmt.Errorf("failed to parse vector: %w", err)
				}

				buf := r.take(int(vec_buf_size))

				log.Println("vector buf string: ", string(buf))
				vec, err := DecodeStateVector([]byte(buf))

				if err != nil {
					return nil, fmt.Errorf("failed to parse vector: %w", err)
				}

				if r.next() != asbyte(VEC_END_SYMBOL) {
					return nil, errors.New(`unterminated ved: missing ']'`)
				}

				op.Args = append(op.Args, vec.Var())

			default:
				return nil, fmt.Errorf("unexpected character after $: %q", r.peek())
			}
		}
	}
}

func ParseCompactOperations(input []byte) ([]*Operation, error) {
	r := &reader{src: input, pos: 0}
	var ops []*Operation

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
