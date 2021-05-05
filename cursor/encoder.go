package cursor

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pilagod/gorm-cursor-paginator/internal/util"
)

// NewEncoder creates cursor encoder
func NewEncoder(keys ...string) *Encoder {
	return &Encoder{keys}
}

// Encoder cursor encoder
type Encoder struct {
	keys []string
}

// Encode encodes model into cursor
func (e *Encoder) Encode(model interface{}) (string, error) {
	b, err := e.marshalJSON(model)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (e *Encoder) marshalJSON(model interface{}) ([]byte, error) {
	rv := util.ReflectValue(model)
	fs := make([]interface{}, len(e.keys))
	for i, key := range e.keys {
		v := util.ReflectValue(rv.FieldByName(key))
		if !v.IsValid() {
			return nil, ErrInvalidModel
		}
		fs[i] = v.Interface()
	}
	result, err := json.Marshal(fs)
	if err != nil {
		return nil, ErrInvalidModel
	}
	return result, nil
}
