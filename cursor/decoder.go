package cursor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
)

var (
	ErrDecodeInvalidCursor = errors.New("invalid cursor for decoding")
	ErrDecodeInvalidModel  = errors.New("invalid model for decoding")
	ErrDecodeKeyUnknown    = errors.New("unknown key on decoded model")
)

// NewDecoder creates cursor decoder for model
func NewDecoder(model interface{}, keys ...string) (*Decoder, error) {
	modelType := reflectType(model)
	// model must be a struct
	if modelType.Kind() != reflect.Struct {
		return nil, ErrDecodeInvalidModel
	}
	for _, key := range keys {
		if _, ok := modelType.FieldByName(key); !ok {
			return nil, ErrDecodeKeyUnknown
		}
	}
	return &Decoder{
		modelType: modelType,
		keys:      keys,
	}, nil
}

type Decoder struct {
	modelType reflect.Type
	keys      []string
}

func (d *Decoder) Decode(cursor string) (fields []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrDecodeInvalidCursor
			return
		}
	}()
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil || !json.Valid(b) {
		return nil, ErrDecodeInvalidCursor
	}
	jd := json.NewDecoder(bytes.NewBuffer(b))
	if t, err := jd.Token(); err != nil || t != json.Delim('[') {
		return nil, ErrDecodeInvalidCursor
	}
	for _, key := range d.keys {
		f, _ := d.modelType.FieldByName(key)
		rt := f.Type
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		v := reflect.New(rt).Interface()
		if err := jd.Decode(&v); err != nil {
			return nil, ErrDecodeInvalidCursor
		}
		fields = append(fields, reflect.ValueOf(v).Elem().Interface())
	}
	if t, err := jd.Token(); err != nil || t != json.Delim(']') {
		return nil, ErrDecodeInvalidCursor
	}
	return
}
