package paginator

func (s *paginatorSuite) TestPaginateNoRule() {
	var orders []order
	_, _, err := New(&Config{
		Rules: []Rule{},
	}).Paginate(s.db, &orders)
	s.Equal(ErrNoRule, err)
}

func (s *paginatorSuite) TestPaginateInvalidLimit() {
	var orders []order
	_, _, err := New(&Config{
		Limit: -1,
	}).Paginate(s.db, &orders)
	s.Equal(ErrInvalidLimit, err)
}

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

func (s *paginatorSuite) TestPaginateInvalidAfterCursor() {
	var orders []order
	_, _, err := New(
		WithAfter("invalid cursor"),
	).Paginate(s.db, &orders)
	s.Equal(ErrInvalidCursor, err)
}

func (s *paginatorSuite) TestPaginateInvalidBeforeCursor() {
	var orders []order
	_, _, err := New(
		WithBefore("invalid cursor"),
	).Paginate(s.db, &orders)
	s.Equal(ErrInvalidCursor, err)
}

func (s *paginatorSuite) TestPaginateInvalidModel() {
	var unknown struct {
		UnknownKey string
	}
	_, _, err := New(
		WithKeys("ID"),
	).Paginate(s.db, &unknown)
	s.Equal(ErrInvalidModel, err)
}
