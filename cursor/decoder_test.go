package cursor

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestDecoder(t *testing.T) {
	suite.Run(t, &decoderSuite{})
}

type decoderSuite struct {
	suite.Suite
}

func (s *decoderSuite) TestModelKeyNotMatched() {
	_, err := NewDecoder("Key").Decode("cursor", struct{ ID string }{})
	s.Equal(ErrInvalidModel, err)
}

func (s *decoderSuite) TestNonStructModel() {
	_, err := NewDecoder("Key").Decode("cursor", 123)
	s.Equal(ErrInvalidModel, err)
}

func (s *decoderSuite) TestInvalidCursorFormat() {
	type model struct {
		Value string
	}
	d := NewDecoder("Value")

	_, err := d.Decode("123", model{})
	s.Equal(ErrInvalidCursor, err)

	c := base64.StdEncoding.EncodeToString([]byte(`{"value": "123"}`))
	_, err = d.Decode(c, model{})
	s.Equal(ErrInvalidCursor, err)

	c = base64.StdEncoding.EncodeToString([]byte(`["123"}`))
	_, err = d.Decode(c, model{})
	s.Equal(ErrInvalidCursor, err)
}

func (s *decoderSuite) TestInvalidCursorType() {
	c, _ := NewEncoder("Value").Encode(struct{ Value int }{123})
	_, err := NewDecoder("Value").Decode(c, struct{ Value string }{})
	s.Equal(ErrInvalidCursor, err)
}
