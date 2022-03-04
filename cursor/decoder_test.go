package cursor

import (
	"encoding/base64"
	"testing"

	"reflect"

	"github.com/stretchr/testify/suite"
)

func TestDecoder(t *testing.T) {
	suite.Run(t, &decoderSuite{})
}

type decoderSuite struct {
	suite.Suite
}

/* decode */

func (s *decoderSuite) TestDecodeKeyNotMatchedModel() {
	_, err := NewDecoder([]string{"Key"}, []*reflect.Type{nil}).Decode("cursor", struct{ ID string }{})
	s.Equal(ErrInvalidModel, err)
}

func (s *decoderSuite) TestDecodeNonStructModel() {
	_, err := NewDecoder([]string{"Key"}, []*reflect.Type{nil}).Decode("cursor", 123)
	s.Equal(ErrInvalidModel, err)
}

func (s *decoderSuite) TestDecodeInvalidCursorFormat() {
	type model struct {
		Value string
	}
	d := NewDecoder([]string{"Value"}, []*reflect.Type{nil})

	// cursor must be a base64 encoded string
	_, err := d.Decode("123", model{})
	s.Equal(ErrInvalidCursor, err)

	// cursor must be a valid json
	c := base64.StdEncoding.EncodeToString([]byte(`["123"}`))
	_, err = d.Decode(c, model{})
	s.Equal(ErrInvalidCursor, err)

	// cursor must be a json array
	c = base64.StdEncoding.EncodeToString([]byte(`{"value": "123"}`))
	_, err = d.Decode(c, model{})
	s.Equal(ErrInvalidCursor, err)
}

func (s *decoderSuite) TestDecodeInvalidCursorType() {
	c, _ := NewEncoder([]string{"Value"}, []interface{}{nil}).Encode(struct{ Value int }{123})
	_, err := NewDecoder([]string{"Value"}, []*reflect.Type{nil}).Decode(c, struct{ Value string }{})
	s.Equal(ErrInvalidCursor, err)
}

/* decode struct */

func (s *decoderSuite) TestDecodeStructInvalidModel() {
	err := NewDecoder([]string{"Value"}, []*reflect.Type{nil}).DecodeStruct("123", struct{ ID string }{})
	s.Equal(ErrInvalidModel, err)
}

func (s *decoderSuite) TestDecodeStructInvalidCursor() {
	err := NewDecoder([]string{"Value"}, []*reflect.Type{nil}).DecodeStruct("123", struct{ Value string }{})
	s.Equal(ErrInvalidCursor, err)
}

/* decode custom types */

func (s *decoderSuite) TestDecodeCustomTypes() {
	type MyType map[string]interface{}

	testCases := []struct {
		name           string
		cursor         string
		typ            reflect.Type
		expectedFields interface{}
	}{
		{
			"nil int",
			`[null]`,
			reflect.PtrTo(reflect.TypeOf(0)),
			[]interface{}{(*int)(nil)},
		},
		{
			"nil string",
			`[null]`,
			reflect.PtrTo(reflect.TypeOf("")),
			[]interface{}{(*string)(nil)},
		},
		{
			"nil float",
			`[null]`,
			reflect.PtrTo(reflect.TypeOf(0.1)),
			[]interface{}{(*float64)(nil)},
		},
		{
			"nil bool",
			`[null]`,
			reflect.PtrTo(reflect.TypeOf(false)),
			[]interface{}{(*bool)(nil)},
		},
		{
			"int",
			`[10]`,
			reflect.TypeOf(0),
			[]interface{}{10},
		},
		{
			"float",
			`[1.5]`,
			reflect.TypeOf(0.0),
			[]interface{}{1.5},
		},
		{
			"string",
			`["A"]`,
			reflect.TypeOf(""),
			[]interface{}{"A"},
		},
		{
			"boolean",
			`[false]`,
			reflect.TypeOf(false),
			[]interface{}{false},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			c := base64.StdEncoding.EncodeToString([]byte(test.cursor))
			fields, err := NewDecoder(
				[]string{"Data"},
				[]*reflect.Type{&test.typ},
			).Decode(c, struct{ Data MyType }{})

			s.Nil(err)
			s.Assert().Equal(test.expectedFields, fields)
		})
	}
}
