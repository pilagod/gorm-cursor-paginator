package paginator

import (
	pc "github.com/pilagod/gorm-cursor-paginator/v2/cursor"
)

// Cursor re-exports cursor.Cursor
type Cursor = pc.Cursor

// CursorCodec encodes/decodes cursor
type CursorCodec interface {
	// Encode encodes model fields into cursor
	Encode(
		fields []pc.EncoderField,
		model interface{},
	) (string, error)

	// Decode decodes cursor into model fields
	Decode(
		fields []pc.DecoderField,
		cursor string,
		model interface{},
	) ([]interface{}, error)
}

// JSONCursorCodec encodes/decodes cursor in JSON format
type JSONCursorCodec struct{}

// Encode encodes model fields into JSON format cursor
func (*JSONCursorCodec) Encode(
	fields []pc.EncoderField,
	model interface{},
) (string, error) {
	return pc.NewEncoder(fields).Encode(model)
}

// Decode decodes JSON format cursor into model fields
func (*JSONCursorCodec) Decode(
	fields []pc.DecoderField,
	cursor string,
	model interface{},
) ([]interface{}, error) {
	return pc.NewDecoder(fields).Decode(cursor, model)
}
