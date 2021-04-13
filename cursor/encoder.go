package cursor

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
)

// NewEncoder creates cursor encoder
func NewEncoder(keys ...string) *Encoder {
	return &Encoder{keys}
}

type Encoder struct {
	keys []string
}

func (e *Encoder) Encode(v interface{}) (string, error) {
	b, err := e.marshalJSON(v)
	if err != nil {
		return "", nil
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (e *Encoder) marshalJSON(value interface{}) ([]byte, error) {
	rv := reflectValue(value)
	fields := make([]interface{}, len(e.keys))
	for i, key := range e.keys {
		v := rv.FieldByName(key)
		// TODO: check is zero
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		fields[i] = v.Interface()
	}
	// @TODO: return proper error
	return json.Marshal(fields)
}
