package paginator

import (
	"github.com/pilagod/gorm-cursor-paginator/cursor"
)

func (s *paginatorSuite) TestPaginateInvalidLimit() {
	var placeholder interface{}
	_, _, err := New(&Config{
		Limit: -1,
	}).Paginate(s.db, &placeholder)
	s.Equal(ErrInvalidLimit, err)
}

func (s *paginatorSuite) TestPaginateInvalidOrder() {
	var placeholder interface{}
	_, _, err := New(&Config{
		Order: "123",
	}).Paginate(s.db, &placeholder)
	s.Equal(ErrInvalidOrder, err)
}

func (s *paginatorSuite) TestPaginateInvalidOrderOnRules() {
	var placeholder interface{}
	_, _, err := New(&Config{
		Rules: []Rule{
			{
				Key:   "ID",
				Order: "123",
			},
		},
	}).Paginate(s.db, &placeholder)
	s.Equal(ErrInvalidOrder, err)
}

func (s *paginatorSuite) TestPaginateInvalidCursor() {
	var orders []order
	_, _, err := New(
		WithAfter("invalid cursor"),
	).Paginate(s.db, &orders)
	s.Equal(cursor.ErrDecodeInvalidCursor, err)
}

func (s *paginatorSuite) TestPaginateUnknownKey() {
	var unknown struct {
		UnknownKey string
	}
	_, _, err := New(
		WithKeys("ID"),
	).Paginate(s.db, &unknown)
	s.Equal(ErrInvalidModel, err)
}

func (s *paginatorSuite) TestPaginateUnknownKeyForDecoding() {
	var unknown struct {
		UnknownKey string
	}
	_, _, err := New(
		WithRules([]Rule{
			{
				Key:     "ID",
				SQLRepr: "id",
			},
		}...),
	).Paginate(s.db, &unknown)
	s.Equal(cursor.ErrDecodeUnknownKey, err)
}
