package cursor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"
)

// CursorDecoder decoder for cursor
type CursorDecoder interface {
	Decode(cursor string) []interface{}
}

// NewCursorDecoder creates cursor decoder
func NewCursorDecoder(ref interface{}, keys ...string) (CursorDecoder, error) {
	// Get the reflected type
	rt := toReflectValue(ref).Type()

	// Reduce reflect type to underlying struct
	for rt.Kind() == reflect.Slice || rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		// element of out must be struct, if not, just pass it to gorm to handle the error
		return nil, ErrInvalidDecodeReference
	}

	return &cursorDecoder{ref: rt, keys: keys}, nil
}

// Errors for decoders
var (
	ErrInvalidDecodeReference = errors.New("decode reference should be struct")
	ErrInvalidField           = errors.New("invalid field")
	ErrInvalidOldField        = errors.New("invalid old field")
)

type cursorDecoder struct {
	// ref is the reference objects reflected type
	ref  reflect.Type
	keys []string
}

func (d *cursorDecoder) Decode(cursor string) []interface{} {
	b, err := base64.StdEncoding.DecodeString(cursor)
	// @TODO: return proper error
	if err != nil {
		return nil
	}

	// If it is not valid JSON, we should attempt to use the old decoding
	// technique for backwards compatability.
	if !json.Valid(b) {
		return decodeOld(b)
	}

	// Create a JSON decoder
	dec := json.NewDecoder(bytes.NewBuffer(b))

	// Read open bracket
	_, err = dec.Token()
	if err != nil {
		return nil
	}

	// Iterate over each key and decode the value
	result := make([]interface{}, len(d.keys))
	for i, key := range d.keys {
		// Find the field in the struct
		field, ok := d.ref.FieldByName(key)
		if !ok {
			return nil
		}

		// Get a copy of the field. JSON decoding requires a pointer but we want
		// to return the same type as that of the referenced object. Therefore
		// capture whether the value is a pointer or not and we will dereference
		// the unmarshalled value before returning it if it is not originally a
		// pointer.
		isPtr := false
		objType := field.Type
		if objType.Kind() == reflect.Ptr {
			isPtr = true
			objType = objType.Elem()
		}
		v := reflect.New(objType).Interface()

		// Decode the value
		if err := dec.Decode(&v); err != nil {
			return nil
		}

		// Need to dereference since everything is now a pointer
		if !isPtr {
			v = reflect.ValueOf(v).Elem().Interface()
		}
		result[i] = v
	}

	return result
}

/* deprecated */

func decodeOld(b []byte) []interface{} {
	fieldsWithType := strings.Split(string(b), ",")
	fields := make([]interface{}, len(fieldsWithType))
	for i, fieldWithType := range fieldsWithType {
		v, err := revert(fieldWithType)
		if err != nil {
			// Failed to parse old encoding
			return nil
		}

		fields[i] = v
	}
	return fields
}

type fieldType string

const (
	fieldString fieldType = "STRING"
	fieldTime   fieldType = "TIME"
)

func revert(fieldWithType string) (interface{}, error) {
	field, fieldType, err := parse(fieldWithType)
	if err != nil {
		return nil, err
	}

	switch fieldType {
	case fieldTime:
		t, err := time.Parse(time.RFC3339Nano, field)
		if err != nil {
			t = time.Now().UTC()
		}
		return t, nil
	default:
		return field, nil
	}
}

func parse(fieldWithType string) (string, fieldType, error) {
	sep := strings.LastIndex(fieldWithType, "?")
	if sep == -1 {
		return "", fieldString, ErrInvalidOldField
	}

	field := fieldWithType[:sep]
	fieldType := fieldType(fieldWithType[sep+1:])
	return field, fieldType, nil
}
