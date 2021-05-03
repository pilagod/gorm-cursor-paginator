package paginator

import "github.com/pilagod/gorm-cursor-paginator/cursor"

func (s *paginatorSuite) TestPaginateInvalidCursor() {
	var orders []order
	_, _, err := New(&Config{
		After: "invalid cursor",
	}).Paginate(s.db, &orders)
	s.Equal(cursor.ErrDecodeInvalidCursor, err)
}

func (s *paginatorSuite) TestPaginateUnknownKey() {
	var unknown struct {
		UnknownKey string
	}
	_, _, err := New(&Config{
		Keys: []string{"ID"},
	}).Paginate(s.db, &unknown)
	s.Equal(cursor.ErrDecodeKeyUnknown, err)
}
