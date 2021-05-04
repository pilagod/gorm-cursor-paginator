package cursor

import "errors"

// Errors for decoder
var (
	ErrDecodeInvalidCursor = errors.New("invalid cursor for decoding")
	ErrDecodeInvalidModel  = errors.New("invalid model for decoding")
	ErrDecodeKeyUnknown    = errors.New("unknown key on decoded model")
)
