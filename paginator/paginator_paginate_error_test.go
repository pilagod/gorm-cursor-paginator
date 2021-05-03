package paginator

import (
	"github.com/pilagod/gorm-cursor-paginator/cursor"
)

func (s *paginatorSuite) TestPaginateInvalidOrder() {
	var orders []order
	_, _, err := New(&Config{
		Order: "123",
	}).Paginate(s.db, &orders)
	s.Equal(ErrInvalidOrder, err)
}

func (s *paginatorSuite) TestPaginateInvalidOrderOnRules() {
	var orders []order
	_, _, err := New(&Config{
		Rules: []Rule{
			{
				Key:   "ID",
				Order: "123",
			},
		},
	}).Paginate(s.db, &orders)
	s.Equal(ErrInvalidOrder, err)
}

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
