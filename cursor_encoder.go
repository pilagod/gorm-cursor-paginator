package paginator

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
)

// CursorEncoder encoder for cursor
type CursorEncoder interface {
	Encode(v interface{}) string
}

// NewCursorEncoder creates cursor encoder
func NewCursorEncoder(keys ...string) CursorEncoder {
	return &cursorEncoder{keys}
}

type cursorEncoder struct {
	keys []string
}

func (e *cursorEncoder) Encode(v interface{}) string {
	return base64.StdEncoding.EncodeToString(e.marshalJSON(v))
}

func (e *cursorEncoder) marshalJSON(value interface{}) []byte {
	rv, ok := value.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(value)
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	fields := make([]interface{}, len(e.keys))
	for i, key := range e.keys {
		fields[i] = rv.FieldByName(key).Interface()
	}
	// @TODO: return proper error
	b, _ := json.Marshal(fields)
	return b
}
