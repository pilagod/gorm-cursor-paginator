package cursor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"

	"github.com/pilagod/gorm-cursor-paginator/internal/util"
)

// NewDecoder creates cursor decoder for model
func NewDecoder(keys ...string) *Decoder {
	return &Decoder{keys}
}

// Decoder cursor decoder
type Decoder struct {
	keys []string
}

// Decode decodes cursor into values
func (d *Decoder) Decode(cursor string, model interface{}) (fields []interface{}, err error) {
	if err = d.validate(model); err != nil {
		return
	}
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
		// key is already validated at validation phase
		f, _ := util.ReflectType(model).FieldByName(key)
		v := reflect.New(util.ReflectType(f.Type)).Interface()
		if err := jd.Decode(&v); err != nil {
			return nil, ErrInvalidCursor
		}
		fields = append(fields, reflect.ValueOf(v).Elem().Interface())
	}
	// cursor must be a valid json after previous checks,
	// so no need to check whether "]" is the last token
	return
}

func (d *Decoder) validate(model interface{}) error {
	modelType := util.ReflectType(model)
	// model must be a struct
	if modelType.Kind() != reflect.Struct {
		return ErrInvalidModel
	}
	for _, key := range d.keys {
		if _, ok := modelType.FieldByName(key); !ok {
			return ErrInvalidModel
		}
	}
	return nil
}
