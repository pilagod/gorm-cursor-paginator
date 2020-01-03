package paginator

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

// CursorDecoder decoder for cursor
type CursorDecoder interface {
	Decode(cursor string) []interface{}
}

// NewCursorDecoder creates cursor decoder
func NewCursorDecoder(model interface{}, keys ...string) CursorDecoder {
	decoder := &cursorDecoder{keys: keys}
	decoder.initKeyKinds(model)
	return decoder
}

type kind uint

const (
	kindInvalid kind = iota
	kindBool
	kindInt
	kindUint
	kindFloat
	kindString
	kindTime
)

var (
	errInvalidFieldType = errors.New("invalid field type")
)

type cursorDecoder struct {
	keys     []string
	keyKinds []kind
}

func (d *cursorDecoder) Decode(cursor string) []interface{} {
	b, err := base64.StdEncoding.DecodeString(cursor)
	// @TODO: return proper error
	if err != nil {
		return nil
	}
	return d.unmarshalJSON(b)
}

func (d *cursorDecoder) initKeyKinds(model interface{}) {
	rv, ok := model.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(model)
	}
	rt := rv.Type()
	for rt.Kind() == reflect.Slice || rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		// element of out must be struct, if not, just pass it to gorm to handle the error
		return
	}
	d.keyKinds = make([]kind, len(d.keys))
	for i, key := range d.keys {
		field, _ := rt.FieldByName(key)
		d.keyKinds[i] = toKind(field.Type)
	}
}

func (d *cursorDecoder) unmarshalJSON(bytes []byte) []interface{} {
	var fields []interface{}
	err := json.Unmarshal(bytes, &fields)
	// @TODO: return proper error
	if err != nil {
		return nil
	}
	return d.castJSONFields(fields)
}

func (d *cursorDecoder) castJSONFields(fields []interface{}) []interface{} {
	result := make([]interface{}, len(fields))
	for i, field := range fields {
		kind := d.keyKinds[i]
		switch f := field.(type) {
		case bool:
			bv, err := castJSONBool(f, kind)
			if err != nil {
				return nil
			}
			result[i] = bv
		case float64:
			fv, err := castJSONFloat(f, kind)
			if err != nil {
				return nil
			}
			result[i] = fv
		case string:
			sv, err := castJSONString(f, kind)
			if err != nil {
				return nil
			}
			result[i] = sv
		default:
			// invalid field
			return nil
		}
	}
	return result
}

func castJSONBool(value bool, kind kind) (interface{}, error) {
	if kind != kindBool {
		return nil, errInvalidFieldType
	}
	return value, nil
}

func castJSONFloat(value float64, kind kind) (interface{}, error) {
	switch kind {
	case kindInt:
		return int(value), nil
	case kindUint:
		return uint(value), nil
	case kindFloat:
		return value, nil
	}
	return nil, errInvalidFieldType
}

func castJSONString(value string, kind kind) (interface{}, error) {
	if kind != kindString && kind != kindTime {
		return nil, errInvalidFieldType
	}
	if kind == kindString {
		return value, nil
	}
	tv, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, errInvalidFieldType
	}
	return tv, nil
}

func toKind(rt reflect.Type) kind {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	// Kind() treats time.Time as struct, so we need specific test for time.Time
	if rt.ConvertibleTo(reflect.TypeOf(time.Time{})) {
		return kindTime
	}
	switch rt.Kind() {
	case reflect.Bool:
		return kindBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return kindInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return kindUint
	case reflect.Float32, reflect.Float64:
		return kindFloat
	default:
		return kindString
	}
}
