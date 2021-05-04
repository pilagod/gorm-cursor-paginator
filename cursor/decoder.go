package cursor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"

	"github.com/pilagod/gorm-cursor-paginator/internal/util"
)

// NewDecoder creates cursor decoder for model
func NewDecoder(model interface{}, keys ...string) (*Decoder, error) {
	modelType := util.ReflectType(model)
	// model must be a struct
	if modelType.Kind() != reflect.Struct {
		return nil, ErrInvalidModel
	}
	// validate keys
	for _, key := range keys {
		if _, ok := modelType.FieldByName(key); !ok {
			return nil, ErrInvalidModel
		}
	}
	return &Decoder{
		modelType: modelType,
		keys:      keys,
	}, nil
}

// Decoder cursor decoder
type Decoder struct {
	modelType reflect.Type
	keys      []string
}

// Decode decodes cursor into values
func (d *Decoder) Decode(cursor string) (values []interface{}, err error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	// ensure cursor content is json
	if err != nil || !json.Valid(b) {
		return nil, ErrInvalidCursor
	}
	jd := json.NewDecoder(bytes.NewBuffer(b))
	// ensure cursor content is json array
	if t, err := jd.Token(); err != nil || t != json.Delim('[') {
		return nil, ErrInvalidCursor
	}
	for _, key := range d.keys {
		// key is already validated when decoder is constructed
		f, _ := d.modelType.FieldByName(key)
		v := reflect.New(util.ReflectType(f.Type)).Interface()
		if err := jd.Decode(&v); err != nil {
			return nil, ErrInvalidCursor
		}
		values = append(values, reflect.ValueOf(v).Elem().Interface())
	}
	// cursor must be a valid json after previous checks,
	// so no need to check whether "]" is the last token
	return
}
