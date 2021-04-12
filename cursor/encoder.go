package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// NewEncoder creates cursor encoder
func NewEncoder(keys ...string) *Encoder {
	return &Encoder{keys}
}

type Encoder struct {
	keys []string
}

func (e *Encoder) Encode(v interface{}) string {
	return base64.StdEncoding.EncodeToString(e.marshalJSON(v))
}

func (e *Encoder) marshalJSON(value interface{}) []byte {
	rv := toReflectValue(value)
	// reduce reflect value to underlying value
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

/* deprecated */

func encodeOld(rv reflect.Value, keys []string) string {
	fields := make([]string, len(keys))
	for index, key := range keys {
		if rv.Kind() == reflect.Ptr {
			fields[index] = convert(reflect.Indirect(rv).FieldByName(key).Interface())
		} else {
			fields[index] = convert(rv.FieldByName(key).Interface())
		}
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(fields, ",")))
}

func convert(field interface{}) string {
	switch field.(type) {
	case time.Time:
		return fmt.Sprintf("%s?%s", field.(time.Time).UTC().Format(time.RFC3339Nano), fieldTime)
	default:
		return fmt.Sprintf("%v?%s", field, fieldString)
	}
}
