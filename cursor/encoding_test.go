package cursor

import (
	"testing"
	"time"

	"errors"
	"reflect"

	"github.com/stretchr/testify/suite"
)

func TestEncoding(t *testing.T) {
	suite.Run(t, &encodingSuite{})
}

type encodingSuite struct {
	suite.Suite
}

/* bool */

type boolModel struct {
	Value    bool
	ValuePtr *bool
}

func (s *encodingSuite) TestBool() {
	c, _ := s.encodeValue(boolModel{Value: true})
	v, _ := s.decodeValue(boolModel{}, c)
	s.Equal(true, v)
}

func (s *encodingSuite) TestBoolPtr() {
	b := true
	c, _ := s.encodeValuePtr(boolModel{ValuePtr: &b})
	v, _ := s.decodeValuePtr(boolModel{}, c)
	s.Equal(true, *(v.(*bool)))
}

/* int */

type intModel struct {
	Value    int
	ValuePtr *int
}

func (s *encodingSuite) TestInt() {
	c, _ := s.encodeValue(intModel{Value: 123})
	v, _ := s.decodeValue(intModel{}, c)
	s.Equal(int(123), v)
}

func (s *encodingSuite) TestIntPtr() {
	i := 123
	c, _ := s.encodeValuePtr(intModel{ValuePtr: &i})
	v, _ := s.decodeValuePtr(intModel{}, c)
	s.Equal(int(123), *(v.(*int)))
}

/* uint */

type uintModel struct {
	Value    uint
	ValuePtr *uint
}

func (s *encodingSuite) TestUint() {
	c, _ := s.encodeValue(uintModel{Value: 123})
	v, _ := s.decodeValue(uintModel{}, c)
	s.Equal(uint(123), v)
}

func (s *encodingSuite) TestUintPtr() {
	ui := uint(123)
	c, _ := s.encodeValuePtr(uintModel{ValuePtr: &ui})
	v, _ := s.decodeValuePtr(uintModel{}, c)
	s.Equal(uint(123), *(v.(*uint)))
}

/* float */
type floatModel struct {
	Value    float64
	ValuePtr *float64
}

func (s *encodingSuite) TestFloat() {
	c, _ := s.encodeValue(floatModel{Value: 123.45})
	v, _ := s.decodeValue(floatModel{}, c)
	s.Equal(float64(123.45), v)
}

func (s *encodingSuite) TestFloatPtr() {
	f := 123.45
	c, _ := s.encodeValuePtr(floatModel{ValuePtr: &f})
	v, _ := s.decodeValuePtr(floatModel{}, c)
	s.Equal(float64(123.45), *(v.(*float64)))
}

/* string */

type stringModel struct {
	Value    string
	ValuePtr *string
}

func (s *encodingSuite) TestString() {
	c, _ := s.encodeValue(stringModel{Value: "hello"})
	v, _ := s.decodeValue(stringModel{}, c)
	s.Equal("hello", v)
}

func (s *encodingSuite) TestStringPtr() {
	str := "hello"
	c, _ := s.encodeValuePtr(stringModel{ValuePtr: &str})
	v, _ := s.decodeValuePtr(stringModel{}, c)
	s.Equal("hello", *(v.(*string)))
}

/* time */

type timeModel struct {
	Value    time.Time
	ValuePtr *time.Time
}

func (s *encodingSuite) TestTime() {
	t := time.Now()
	c, _ := s.encodeValue(timeModel{Value: t})
	v, _ := s.decodeValue(timeModel{}, c)
	s.Equal(t.Second(), v.(time.Time).Second())
}

func (s *encodingSuite) TestTimePtr() {
	t := time.Now()
	c, _ := s.encodeValuePtr(timeModel{ValuePtr: &t})
	v, _ := s.decodeValuePtr(timeModel{}, c)
	s.Equal(t.Second(), v.(*time.Time).Second())
}

/* struct */

type structModel struct {
	Value    structValue
	ValuePtr *structValue
}

type structValue struct {
	Value []byte
}

func (s *encodingSuite) TestStruct() {
	c, _ := s.encodeValue(structModel{
		Value: structValue{Value: []byte("123")},
	})
	v, _ := s.decodeValue(structModel{}, c)
	s.Equal(structValue{Value: []byte("123")}, v)
}

func (s *encodingSuite) TestStructPtr() {
	sv := structValue{Value: []byte("123")}
	c, _ := s.encodeValuePtr(structModel{ValuePtr: &sv})
	v, _ := s.decodeValuePtr(structModel{}, c)
	s.Equal(sv, *(v.(*structValue)))
}

/* multiple */

type multipleModel struct {
	ID        int
	Name      string
	CreatedAt *time.Time
}

func (multipleModel) DecoderFields() []DecoderField {
	return []DecoderField{
		{Key: "ID"},
		{Key: "Name"},
		{Key: "CreatedAt"},
	}
}

func (multipleModel) EncoderFields() []EncoderField {
	return []EncoderField{
		{Key: "ID"},
		{Key: "Name"},
		{Key: "CreatedAt"},
	}
}

func (s *encodingSuite) TestMultipleFields() {
	encoderFields := multipleModel{}.EncoderFields()
	decoderFields := multipleModel{}.DecoderFields()

	t := time.Now()
	c, err := NewEncoder(encoderFields).Encode(multipleModel{
		ID:        123,
		Name:      "Hello",
		CreatedAt: &t,
	})
	s.Nil(err)

	fields, err := NewDecoder(decoderFields).Decode(c, multipleModel{})
	s.Nil(err)

	s.Len(fields, 3)
	s.Equal(123, fields[0])
	s.Equal("Hello", fields[1])
	s.Equal(t.Second(), fields[2].(*time.Time).Second())
}

func (s *encoderSuite) TestMultipleFieldsWithZeroValue() {
	encoderFields := multipleModel{}.EncoderFields()
	decoderFields := multipleModel{}.DecoderFields()

	c, err := NewEncoder(encoderFields).Encode(multipleModel{})
	s.Nil(err)

	fields, err := NewDecoder(decoderFields).Decode(c, multipleModel{})
	s.Nil(err)

	s.Equal(0, fields[0])
	s.Equal("", fields[1])
	s.Equal((*time.Time)(nil), fields[2])
}

/* decode struct */

func (s *encodingSuite) TestMultipleFieldsToStruct() {
	encoderFields := multipleModel{}.EncoderFields()
	decoderFields := multipleModel{}.DecoderFields()

	t := time.Now()
	c, err := NewEncoder(encoderFields).Encode(multipleModel{
		ID:        123,
		Name:      "Hello",
		CreatedAt: &t,
	})
	s.Nil(err)

	var model multipleModel
	err = NewDecoder(decoderFields).DecodeStruct(c, &model)
	s.Nil(err)

	s.Equal(123, model.ID)
	s.Equal("Hello", model.Name)
	s.Equal(t.Second(), (*model.CreatedAt).Second())
}

func (s *encoderSuite) TestMultipleFieldsToStructWithZeroValue() {
	encoderFields := multipleModel{}.EncoderFields()
	decoderFields := multipleModel{}.DecoderFields()

	c, err := NewEncoder(encoderFields).Encode(multipleModel{})
	s.Nil(err)

	var model multipleModel
	err = NewDecoder(decoderFields).DecodeStruct(c, &model)
	s.Nil(err)

	s.Equal(0, model.ID)
	s.Equal("", model.Name)
	s.Equal((*time.Time)(nil), model.CreatedAt)
}

func (s *encodingSuite) encodeValue(v interface{}) (string, error) {
	return NewEncoder([]EncoderField{{Key: "Value"}}).Encode(v)
}

func (s *encodingSuite) encodeValuePtr(v interface{}) (string, error) {
	return NewEncoder([]EncoderField{{Key: "ValuePtr"}}).Encode(v)
}

func (s *encodingSuite) decodeValue(m interface{}, c string) (interface{}, error) {
	fields, err := NewDecoder([]DecoderField{{Key: "Value"}}).Decode(c, m)
	if err != nil {
		return nil, err
	}
	if len(fields) != 1 {
		s.FailNow("invalid value model: %v, fields %v", m, fields)
	}
	return fields[0], nil
}

func (s *encodingSuite) decodeValuePtr(m interface{}, c string) (interface{}, error) {
	fields, err := NewDecoder([]DecoderField{{Key: "ValuePtr"}}).Decode(c, m)
	if err != nil {
		return nil, err
	}
	if len(fields) != 1 {
		s.FailNow("invalid value model: %v, fields %v", m, fields)
	}
	return fields[0], nil
}

/* Custom Types Encoding Decoding */

type MyJSON map[string]interface{}

var MyJSONError = errors.New("meta should be string")

func (t MyJSON) GetCustomTypeValue(meta interface{}) (interface{}, error) {
	key, ok := meta.(string)
	if !ok {
		return nil, MyJSONError
	}
	return t[key], nil
}

func (s *encodingSuite) TestEncodeDecodeCustomTypes() {
	testCases := []struct {
		name  string
		typ   reflect.Type
		value interface{}
	}{
		{
			"nil int",
			reflect.PtrTo(reflect.TypeOf(0)),
			(*int)(nil),
		},
		{
			"nil float",
			reflect.PtrTo(reflect.TypeOf(0.0)),
			(*float64)(nil),
		},
		{
			"nil string",
			reflect.PtrTo(reflect.TypeOf("")),
			(*string)(nil),
		},
		{
			"nil bool",
			reflect.PtrTo(reflect.TypeOf(true)),
			(*bool)(nil),
		},
		{
			"int",
			reflect.TypeOf(0),
			10,
		},
		{
			"float",
			reflect.TypeOf(0.0),
			1.5,
		},
		{
			"string",
			reflect.TypeOf(""),
			"A",
		},
		{
			"boolean",
			reflect.TypeOf(false),
			false,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			// encode value
			c, err := NewEncoder([]EncoderField{
				{Key: "Data", Meta: "key"},
			}).Encode(struct{ Data MyJSON }{MyJSON{"key": test.value}})
			s.Nil(err)

			// decode value
			v, err := NewDecoder([]DecoderField{
				{Key: "Data", Type: &test.typ},
			}).Decode(c, struct{ Data MyJSON }{})
			s.Nil(err)

			// make sure they match
			s.Equal(test.value, v[0])
		})
	}
}
