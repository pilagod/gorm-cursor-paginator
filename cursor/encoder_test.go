package cursor

import (
	"testing"

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

func (s *encoderSuite) TestGetCustomTypeValueError() {
	// meta should be string for MyJSON; hence -1 will error
	e := NewEncoder([]EncoderField{
		{Key: "Data", Meta: -1},
	})
	_, err := e.Encode(struct{ Data MyJSON }{MyJSON{"key": "value"}})
	s.Equal(MyJSONError, err)
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
