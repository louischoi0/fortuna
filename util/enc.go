package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"fortuna/structure"
)

func EncodeInt64Array(arr []int64) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, arr)
	return buf.Bytes()
}

func DecodeInt64Array(data []byte) ([]int64, error) {
	if len(data)%8 != 0 {
		return nil, fmt.Errorf("invalid int64 array bytes: len=%d not multiple of 8", len(data))
	}
	n := len(data) / 8
	result := make([]int64, n)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func EncodeInt64(n int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	return buf
}

func DecodeInt64(b []byte) (int64, error) {
	if len(b) != 8 {
		return 0, errors.New("bytes must has length 8")
	}
	num := int64(binary.BigEndian.Uint64(b))
	return num, nil
}

func EncodeInt32(value int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeInt32(data []byte) (int32, error) {
	var value int32
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func EncodeUint32(n uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, n)
	return buf
}

func DecodeUint32(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, fmt.Errorf("invalid byte length for uint32: expected 4, got %d", len(b))
	}
	return binary.BigEndian.Uint32(b), nil
}

const (
	objectStartFlag byte = 0x01

	typeString = 0
	typeInt64  = 1
	typeStrArr = 2
	typeIntArr = 3
	typeObject = 4
)

var le = binary.LittleEndian

func EncodeOrderedMap(obj *structure.OrderedMap) ([]byte, error) {
	body, err := encodeObjectBody(obj)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := out.WriteByte(objectStartFlag); err != nil {
		return nil, err
	}
	if err := binary.Write(&out, le, uint64(len(body))); err != nil {
		return nil, err
	}
	if _, err := out.Write(body); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func encodeObjectBody(obj *structure.OrderedMap) ([]byte, error) {
	var buf bytes.Buffer

	keys := obj.Keys()

	if err := binary.Write(&buf, le, uint32(len(keys))); err != nil {
		return nil, err
	}

	for _, k := range keys {
		v, _ := obj.Get(k)

		// key
		if err := binary.Write(&buf, le, uint32(len(k))); err != nil {
			return nil, err
		}
		if _, err := buf.WriteString(k); err != nil {
			return nil, err
		}

		// value
		switch x := v.(type) {
		case string:
			if err := writeU8(&buf, typeString); err != nil {
				return nil, err
			}
			b := []byte(x)
			if err := writeU64(&buf, uint64(len(b))); err != nil {
				return nil, err
			}
			if _, err := buf.Write(b); err != nil {
				return nil, err
			}

		case int:
			if err := writeU8(&buf, typeInt64); err != nil {
				return nil, err
			}
			if err := writeU64(&buf, 8); err != nil {
				return nil, err
			}
			if err := binary.Write(&buf, le, int64(x)); err != nil {
				return nil, err
			}

		case int64:
			if err := writeU8(&buf, typeInt64); err != nil {
				return nil, err
			}
			if err := writeU64(&buf, 8); err != nil {
				return nil, err
			}
			if err := binary.Write(&buf, le, x); err != nil {
				return nil, err
			}

		case []string:
			if err := writeU8(&buf, typeStrArr); err != nil {
				return nil, err
			}
			if err := writeU64(&buf, uint64(len(x))); err != nil {
				return nil, err
			}
			for _, s := range x {
				b := []byte(s)
				if err := writeU64(&buf, uint64(len(b))); err != nil {
					return nil, err
				}
				if _, err := buf.Write(b); err != nil {
					return nil, err
				}
			}

		case *structure.OrderedMap:
			if err := writeU8(&buf, typeObject); err != nil {
				return nil, err
			}
			enc, err := EncodeOrderedMap(x)
			if err != nil {
				return nil, err
			}
			if err := writeU64(&buf, uint64(len(enc))); err != nil {
				return nil, err
			}
			if _, err := buf.Write(enc); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unsupported type for key %q: %T", k, v)
		}
	}

	return buf.Bytes(), nil
}

func DecodeOrderedMap(b []byte) (*structure.OrderedMap, error) {
	r := &reader{b: b}

	// flag
	flg, err := r.u8()
	if err != nil {
		return nil, err
	}
	if flg != objectStartFlag {
		return nil, errors.New("invalid object start flag")
	}

	// totalLen
	totalLen, err := r.u64()
	if err != nil {
		return nil, err
	}
	body, err := r.bytes(int(totalLen)) // body 슬라이스 확보
	if err != nil {
		return nil, err
	}

	// body 전용 리더에서 파싱
	br := &reader{b: body}
	keyCount, err := br.u32()
	if err != nil {
		return nil, err
	}

	obj := structure.NewOrderedMap()

	for i := uint32(0); i < keyCount; i++ {
		klen, err := br.u32()
		if err != nil {
			return nil, err
		}
		kb, err := br.bytes(int(klen))
		if err != nil {
			return nil, err
		}
		key := string(kb)

		typ, err := br.u8()
		if err != nil {
			return nil, err
		}
		vlen, err := br.u64()
		if err != nil {
			return nil, err
		}

		switch typ {
		case typeString:
			vb, err := br.bytes(int(vlen))
			if err != nil {
				return nil, err
			}
			obj.Set(key, string(vb))

		case typeInt64:
			if vlen != 8 {
				return nil, fmt.Errorf("invalid int64 length for key %q: %d", key, vlen)
			}
			n, err := br.i64()
			if err != nil {
				return nil, err
			}
			obj.Set(key, n)

		case typeStrArr:
			cnt := vlen
			arr := make([]string, 0, cnt)
			for j := uint64(0); j < cnt; j++ {
				elLen, err := br.u64()
				if err != nil {
					return nil, err
				}
				elBytes, err := br.bytes(int(elLen))
				if err != nil {
					return nil, err
				}
				arr = append(arr, string(elBytes))
			}
			obj.Set(key, arr)

		case typeIntArr:
			cnt := vlen
			arr := make([]int64, 0, cnt)
			for j := uint64(0); j < cnt; j++ {
				n, err := br.i64()
				if err != nil {
					return nil, err
				}
				arr = append(arr, n)
			}
			obj.Set(key, arr)

		case typeObject:
			subBlob, err := br.bytes(int(vlen)) // 서브 오브젝트 전체 블롭
			if err != nil {
				return nil, err
			}
			m, err := DecodeOrderedMap(subBlob)
			if err != nil {
				return nil, err
			}
			obj.Set(key, m)

		default:
			return nil, fmt.Errorf("unknown value type %d for key %q", typ, key)
		}
	}

	if br.p != len(br.b) {
		return nil, errors.New("object body has trailing bytes")
	}

	return obj, nil
}

// --- helpers ---

func writeU8(buf *bytes.Buffer, v int) error     { return buf.WriteByte(byte(v)) }
func writeU64(buf *bytes.Buffer, v uint64) error { return binary.Write(buf, le, v) }

type reader struct {
	b []byte
	p int
}

func (r *reader) need(n int) error {
	if r.p+n > len(r.b) {
		return errors.New("buffer underflow")
	}
	return nil
}
func (r *reader) u8() (byte, error) {
	if err := r.need(1); err != nil {
		return 0, err
	}
	v := r.b[r.p]
	r.p++
	return v, nil
}
func (r *reader) u32() (uint32, error) {
	if err := r.need(4); err != nil {
		return 0, err
	}
	v := le.Uint32(r.b[r.p : r.p+4])
	r.p += 4
	return v, nil
}
func (r *reader) u64() (uint64, error) {
	if err := r.need(8); err != nil {
		return 0, err
	}
	v := le.Uint64(r.b[r.p : r.p+8])
	r.p += 8
	return v, nil
}
func (r *reader) i64() (int64, error) {
	u, err := r.u64()
	return int64(u), err
}
func (r *reader) bytes(n int) ([]byte, error) {
	if err := r.need(n); err != nil {
		return nil, err
	}
	v := r.b[r.p : r.p+n]
	r.p += n
	return v, nil
}

func PadLeftS(a string, length int) string {
	n := len(a)
	if n > length {
		return a
	}
	if n == length {
		return a
	}

	buf := make([]byte, length)
	copy(buf[length-n:], []byte(a))
	return string(buf)
}

func UnpadLeftS(s string) string {
	b := []byte(s)
	i := 0
	for i < len(b) && b[i] == 0x00 {
		i++
	}
	return string(b[i:])
}
