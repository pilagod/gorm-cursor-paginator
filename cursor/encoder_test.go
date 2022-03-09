package cursor

import (
	"testing"

	"encoding/base64"

	"github.com/stretchr/testify/suite"
)

func TestEncoder(t *testing.T) {
	suite.Run(t, &encoderSuite{})
}

type encoderSuite struct {
	suite.Suite
}

func (s *encoderSuite) TestInvalidModel() {
	e := NewEncoder([]EncoderField{{Key: "ID"}})
	_, err := e.Encode(struct{}{})
	s.Equal(ErrInvalidModel, err)
}

func (s *encoderSuite) TestInvalidModelFieldType() {
	// https://stackoverflow.com/questions/33903552/what-input-will-cause-golangs-json-marshal-to-return-an-error
	e := NewEncoder([]EncoderField{{Key: "ID"}})
	_, err := e.Encode(
		struct {
			ID chan int
		}{make(chan int)},
	)
	s.Equal(ErrInvalidModel, err)
}

func (s *encoderSuite) TestZeroValue() {
	e := NewEncoder([]EncoderField{{Key: "ID"}})
	_, err := e.Encode(struct{ ID string }{})
	s.Nil(err)
}

func (s *encoderSuite) TestZeroValuePtr() {
	e := NewEncoder([]EncoderField{{Key: "ID"}})
	_, err := e.Encode(struct{ ID *string }{})
	s.Nil(err)
}

/* encode custom types */

type MyType map[string]interface{}

func (t MyType) GetCustomTypeValue(meta interface{}) interface{} {
	return t[meta.(string)]
}

func (s *decoderSuite) TestEncodeCustomTypes() {
	testCases := []struct {
		name           string
		value          interface{}
		expectedCursor string
	}{
		{
			"nil",
			nil,
			"[null]",
		},
		{
			"int",
			10,
			"[10]",
		},
		{
			"float",
			0.5,
			"[0.5]",
		},
		{
			"string",
			"a",
			"[\"a\"]",
		},
		{
			"boolean",
			false,
			"[false]",
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {

			e := NewEncoder([]EncoderField{{Key: "Data", Meta: "key"}})
			c, err := e.Encode(struct{ Data MyType }{MyType{"key": test.value}})
			s.Nil(err)

			// decode cursor
			returnedCursor, err := base64.StdEncoding.DecodeString(c)
			s.Nil(err)

			s.Assert().Equal(test.expectedCursor, string(returnedCursor))
		})
	}
}
