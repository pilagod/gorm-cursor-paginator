package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestCursor(t *testing.T) {
	suite.Run(t, &cursorSuite{})
}

/* suite */

type cursorSuite struct {
	suite.Suite
}

/* suite test cases */

func (s *cursorSuite) TestCursorEncoderAndDecoder() {
	var model = createCursorModelFixture()
	cursor := model.Encode()
	fields, _ := model.Decode(cursor)
	s.assertFields(model, fields)
}

/* cursor encoder */

func (s *cursorSuite) TestCursorEncoderBackwardCompatibility() {
	var model = createCursorModelFixture()
	cursor := model.Encode()
	fields := Decode(cursor)
	s.assertDeprecatedFields(model, fields)
}

/* cursor decoder */

func (s *cursorSuite) TestCursorDecoderShouldReturnErrorWhenRefIsNotStruct() {
	var nonStructRef int
	_, err := NewDecoder(nonStructRef, "Key")
	s.Equal(ErrInvalidDecodeReference, err)
}

func (s *cursorSuite) TestCursorDecoderShouldReturnNilWhenCursorIsNotBase64Encoded() {
	var model = createCursorModelFixture()
	decoder, _ := model.Decoder()
	fields := decoder.Decode("hello world")
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDecoderShouldReturnNilWhenBoolValueIsNotMatched() {
	var model = createCursorModelFixture()
	cursor := model.EncodeReplace("Bool", 123)
	fields, _ := model.Decode(cursor)
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDecoderShouldReturnNilWhenIntValueIsNotMatched() {
	var model = createCursorModelFixture()
	cursor := model.EncodeReplace("Int", "123")
	fields, _ := model.Decode(cursor)
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDecoderShouldReturnNilWhenUintValueIsNotMatched() {
	var model = createCursorModelFixture()
	cursor := model.EncodeReplace("Uint", "123")
	fields, _ := model.Decode(cursor)
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDecoderShouldReturnNilWhenFloatValueIsNotMatched() {
	var model = createCursorModelFixture()
	cursor := model.EncodeReplace("Float", "123")
	fields, _ := model.Decode(cursor)
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDecoderShouldReturnNilWhenStringValueIsNotMatched() {
	var model = createCursorModelFixture()
	cursor := model.EncodeReplace("String", true)
	fields, _ := model.Decode(cursor)
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDecoderShouldReturnNilWhenTimeValueIsNotMatched() {
	var model = createCursorModelFixture()
	cursor := model.EncodeReplace("Time", "2020 Sun Jan 5 17:51:12")
	fields, _ := model.Decode(cursor)
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDecoderBackwardCompatibility() {
	var model = createCursorModelFixture()
	cursor := model.EncodeOld()
	fields, _ := model.Decode(cursor)
	s.assertDeprecatedFields(model, fields)
}

func (s *cursorSuite) TestCursorDecoderBackwardCompatibilityForPtr() {
	var model = createCursorModelFixture()
	cursor := model.EncodeOldPtr()
	fields, _ := model.Decode(cursor)
	s.assertDeprecatedFields(model, fields)
}

/* cursor deprecated encode & decode */

func (s *cursorSuite) TestCursorDeprecatedEncodeAndDecode() {
	var model = createCursorModelFixture()
	cursor := Encode(reflect.ValueOf(model), model.Keys())
	fields := Decode(cursor)
	s.assertDeprecatedFields(model, fields)
}

func (s *cursorSuite) TestCursorDeprecatedDecodeShouldReturnNilWhenCursorIsNotBase64Encoded() {
	fields := Decode("hello world")
	s.Nil(fields)
}

func (s *cursorSuite) TestCursorDeprecatedEncodeForwardCompatibility() {
	var model = createCursorModelFixture()
	expected := model.Encode()
	got := Encode(reflect.ValueOf(model), model.Keys())
	s.Equal(expected, got)
}

func (s *cursorSuite) TestCursorDeprecatedDecodeBackwardCompatibility() {
	var model = createCursorModelFixture()
	cursor := model.EncodeOld()
	fields := Decode(cursor)
	s.assertDeprecatedFields(model, fields)
}

/* cursor test model */

type cursorModel struct {
	Bool           bool
	Int            int
	Uint           uint
	Float          float64
	String         string
	Time           time.Time
	StructField    structField
	StructFieldPtr *structField
}

type structField struct {
	Value []byte
}

func createCursorModelFixture() cursorModel {
	return cursorModel{
		Bool:   true,
		Int:    1,
		Uint:   2,
		Float:  3.14,
		String: "hello",
		Time:   time.Now(),
		StructField: structField{
			Value: []byte{'t', 'e', 's', 't'},
		},
		StructFieldPtr: &structField{
			Value: []byte{'t', 'e', 's', 't', '2'},
		},
	}
}

func (m *cursorModel) FieldCount() int {
	return len(m.Keys())
}

func (m *cursorModel) Keys() []string {
	return []string{"Bool", "Int", "Uint", "Float", "String", "Time", "StructField", "StructFieldPtr"}
}

func (m *cursorModel) Encode() string {
	return m.Encoder().Encode(m)
}

func (m *cursorModel) Encoder() *Encoder {
	return NewEncoder(m.Keys()...)
}

func (m *cursorModel) EncodeReplace(key string, value interface{}) string {
	b, err := base64.StdEncoding.DecodeString(m.Encode())
	if err != nil {
		panic("cursor encoded from CursorEncoder should be base64 encoded")
	}

	bv, err := json.Marshal(reflect.ValueOf(*m).FieldByName(key).Interface())
	if err != nil {
		panic(err.Error())
	}
	old := string(bv)

	bv, err = json.Marshal(value)
	if err != nil {
		panic(err.Error())
	}
	new := string(bv)

	b = []byte(strings.Replace(string(b), old, new, 1))
	return base64.StdEncoding.EncodeToString(b)
}

func (m *cursorModel) EncodeOld() string {
	return encodeOld(reflect.ValueOf(*m), m.Keys())
}

func (m *cursorModel) EncodeOldPtr() string {
	return encodeOld(reflect.ValueOf(m), m.Keys())
}

func (m *cursorModel) Decode(cursor string) ([]interface{}, error) {
	decoder, err := m.Decoder()
	if err != nil {
		return nil, err
	}
	return decoder.Decode(cursor), nil
}

func (m *cursorModel) Decoder() (*Decoder, error) {
	return NewDecoder(m, m.Keys()...)
}

/* util */

func (s *cursorSuite) assertFields(model cursorModel, fields []interface{}) {
	s.Len(fields, model.FieldCount())

	boolVal, _ := fields[0].(bool)
	s.Equal(model.Bool, boolVal)

	intVal, _ := fields[1].(int)
	s.Equal(model.Int, intVal)

	uintVal, _ := fields[2].(uint)
	s.Equal(model.Uint, uintVal)

	floatVal, _ := fields[3].(float64)
	s.Equal(model.Float, floatVal)

	stringVal, _ := fields[4].(string)
	s.Equal(model.String, stringVal)

	timeVal, _ := fields[5].(time.Time)
	s.assertTime(model.Time, timeVal)

	embeddedVal, _ := fields[6].(structField)
	s.Equal(model.StructField.Value, embeddedVal.Value)

	embeddedPtrVal, _ := fields[7].(*structField)
	s.Equal(model.StructFieldPtr.Value, embeddedPtrVal.Value)
}

func (s *cursorSuite) assertDeprecatedFields(model cursorModel, fields []interface{}) {
	s.Len(fields, model.FieldCount())

	boolVal, _ := fields[0].(string)
	s.Equal(toString(model.Bool), boolVal)

	intVal, _ := fields[1].(string)
	s.Equal(toString(model.Int), intVal)

	uintVal, _ := fields[2].(string)
	s.Equal(toString(model.Uint), uintVal)

	floatVal, _ := fields[3].(string)
	s.Equal(toString(model.Float), floatVal)

	stringVal, _ := fields[4].(string)
	s.Equal(model.String, stringVal)

	timeVal, _ := fields[5].(time.Time)
	s.assertTime(model.Time, timeVal)
}

func (s *cursorSuite) assertTime(expected time.Time, got time.Time) {
	s.Equal(expected.Unix(), got.Unix())
}

func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
