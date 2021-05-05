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

// Decode decodes cursor into values (without pointer) by referencing field type on model.
//
// For example:
//
//  type model struct{
//      Value *int
//  }
//
// will be decoded as []int
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
		// key is already validated at beginning
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

// DecodeStruct decodes cursor into model, model must be a pointer to struct or it will panic.
func (d *Decoder) DecodeStruct(cursor string, model interface{}) error {
	fields, err := d.Decode(cursor, model)
	if err != nil {
		return err
	}
	e := reflect.ValueOf(model).Elem()
	for i, key := range d.keys {
		var v reflect.Value
		f := e.FieldByName(key)
		if f.Kind() == reflect.Ptr {
			v = reflect.New(reflect.ValueOf(fields[i]).Type())
			v.Elem().Set(reflect.ValueOf(fields[i]))
		} else {
			v = reflect.ValueOf(fields[i])
		}
		f.Set(v)
	}
	return nil
}

func (d *Decoder) validate(model interface{}) error {
	modelType := util.ReflectType(model)
	// model's underlying type must be a struct
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
