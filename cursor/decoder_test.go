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

func (s *decoderSuite) TestKeyUnknown() {
	_, err := NewDecoder(struct{}{}, "Hello")
	s.Equal(ErrDecodeKeyUnknown, err)
}

func (s *decoderSuite) TestInvalidModel() {
	_, err := NewDecoder(123, "")
	s.Equal(ErrDecodeInvalidModel, err)
}

func (s *decoderSuite) TestInvalidCursorFormat() {
	d, _ := NewDecoder(struct{ Value string }{}, "Value")

	_, err := d.Decode("123")
	s.Equal(ErrDecodeInvalidCursor, err)

	var c string

	c = base64.StdEncoding.EncodeToString([]byte(`{"value": "123"}`))
	_, err = d.Decode(c)
	s.Equal(ErrDecodeInvalidCursor, err)

	c = base64.StdEncoding.EncodeToString([]byte(`["123"}`))
	_, err = d.Decode(c)
	s.Equal(ErrDecodeInvalidCursor, err)
}

func (s *decoderSuite) TestInvalidCursorType() {
	d, _ := NewDecoder(struct{ Value string }{}, "Value")
	c, _ := NewEncoder("Value").Encode(struct{ Value int }{
		Value: 123,
	})
	_, err := d.Decode(c)
	s.Equal(ErrDecodeInvalidCursor, err)
}
