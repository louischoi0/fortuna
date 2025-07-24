package vm

import "strconv"

type Parameter struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`

	MaxLength int `json:"max_length"`
	MinLength int `json:"min_length"`

	MaxValue int `json:"max_value"`
	MinValue int `json:"min_value"`
}

func (p *Parameter) Validate() bool {
	switch p.Value.(type) {
	case string:
		if p.MaxLength > 0 && len(p.Value.(string)) > p.MaxLength {
			return false
		}
	case int, int8, int16, int32, int64, float32, float64:
		if p.MaxValue > 0 && p.Value.(int) > p.MaxValue {
			return false
		}
	}

	return true
}

func (p *Parameter) AsString() string {
	switch p.Value.(type) {
	case string:
		return p.Value.(string)
	case int, int8, int16, int32, int64, float32, float64:
		return strconv.FormatFloat(p.Value.(float64), 'f', -1, 64)
	default:
		return ""
	}
}

func (p *Parameter) AsInt() int {
	switch p.Value.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return p.Value.(int)
	case string:
		num, err := strconv.Atoi(p.Value.(string))
		if err != nil {
			return 0
		}
		return num
	default:
		return 0
	}
}

func (p *Parameter) AsInt64() int64 {
	switch p.Value.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return p.Value.(int64)
	case string:
		num, err := strconv.ParseInt(p.Value.(string), 10, 64)
		if err != nil {
			return 0
		}
		return num
	default:
		return 0
	}
}
