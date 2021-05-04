package cursor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
)

// NewDecoder creates cursor decoder for model
func NewDecoder(model interface{}, keys ...string) (*Decoder, error) {
	modelType := reflectType(model)
	// model must be a struct
	if modelType.Kind() != reflect.Struct {
		return nil, ErrDecodeInvalidModel
	}
	// validate keys
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
	b, err := base64.StdEncoding.DecodeString(cursor)
	// ensure cursor content is json
	if err != nil || !json.Valid(b) {
		return nil, ErrDecodeInvalidCursor
	}
	jd := json.NewDecoder(bytes.NewBuffer(b))
	// ensure cursor content is json array
	if t, err := jd.Token(); err != nil || t != json.Delim('[') {
		return nil, ErrDecodeInvalidCursor
	}
	for _, key := range d.keys {
		// key is already validated when decoder is constructed
		f, _ := d.modelType.FieldByName(key)
		v := reflect.New(reflectType(f.Type)).Interface()
		if err := jd.Decode(&v); err != nil {
			return nil, ErrDecodeInvalidCursor
		}
		fields = append(fields, reflect.ValueOf(v).Elem().Interface())
	}
	// cursor must be a valid json after previous checks,
	// so no need to check "]" is the last token
	return
}
