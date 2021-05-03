package paginator

import (
	"time"

	"github.com/pilagod/gorm-cursor-paginator/cursor"
)

func (s *paginatorSuite) TestPaginateDefaultOptions() {
	s.givenOrders(12)

	// Default Options
	// * Key: ID
	// * Limit: 10
	// * Order: DESC

	var p1 []order
	_, c, _ := New().Paginate(s.db, &p1)
	s.assertIDRange(p1, 12, 3)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(&Config{
		After: *c.After,
	}).Paginate(s.db, &p2)
	s.assertIDRange(p2, 2, 1)
	s.assertBackwardOnly(c)

	var p3 []order
	_, c, _ = New(&Config{
		Before: *c.Before,
	}).Paginate(s.db, &p3)
	s.assertIDRange(p3, 12, 3)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateSlicePtrs() {
	s.givenOrders(12)

	var p1 []*order
	_, c, _ := New().Paginate(s.db, &p1)
	s.assertIDRange(p1, 12, 3)
	s.assertForwardOnly(c)

	var p2 []*order
	_, c, _ = New(&Config{
		After: *c.After,
	}).Paginate(s.db, &p2)
	s.assertIDRange(p2, 2, 1)
	s.assertBackwardOnly(c)

	var p3 []*order
	_, c, _ = New(&Config{
		Before: *c.Before,
	}).Paginate(s.db, &p3)
	s.assertIDRange(p3, 12, 3)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateNonSlice() {
	s.givenOrders(3)

	var o order
	_, c, _ := New().Paginate(s.db, &o)
	s.Equal(3, o.ID)
	s.assertNoMore(c)
}

func (s *paginatorSuite) TestPaginateNoMore() {
	s.givenOrders(3)

	var orders []order
	_, c, _ := New().Paginate(s.db, &orders)
	s.assertIDRange(orders, 3, 1)
	s.assertNoMore(c)
}

func (s *paginatorSuite) TestPaginateForwardShouldTakePrecedenceOverBackward() {
	s.givenOrders(30)

	var p1 []order
	_, c, _ := New().Paginate(s.db, &p1)
	s.assertIDRange(p1, 30, 21)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(&Config{
		After: *c.After,
	}).Paginate(s.db, &p2)
	s.assertIDRange(p2, 20, 11)
	s.assertBothDirections(c)

	var p3 []order
	_, c, _ = New(&Config{
		After:  *c.After,
		Before: *c.Before,
	}).Paginate(s.db, &p3)
	s.assertIDRange(p3, 10, 1)
	s.assertBackwardOnly(c)
}

func (s *paginatorSuite) TestPaginateSingleKey() {
	// ID ordered by CreatedAt desc: 1, 3, 2
	s.givenOrders([]order{
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now().Add(-1 * time.Hour)},
		{CreatedAt: time.Now()},
	})

	cfg := Config{
		Keys:  []string{"CreatedAt"},
		Limit: 2,
	}

	var p1 []order
	_, c, _ := New(&cfg).Paginate(s.db, &p1)
	s.assertIDs(p1, 1, 3)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(
		&cfg,
		WithAfter(*c.After),
	).Paginate(s.db, &p2)
	s.assertIDs(p2, 2)
	s.assertBackwardOnly(c)

	var p3 []order
	_, c, _ = New(
		&cfg,
		WithBefore(*c.Before),
	).Paginate(s.db, &p3)
	s.assertIDs(p3, 1, 3)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateMultipleKeys() {
	// ID ordered by CreatedAt, ID desc: 2, 3, 1
	s.givenOrders([]order{
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now()},
	})

	cfg := Config{
		Keys:  []string{"CreatedAt", "ID"},
		Limit: 2,
	}

	var p1 []order
	_, c, _ := New(&cfg).Paginate(s.db, &p1)
	s.assertIDs(p1, 2, 3)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(
		&cfg,
		WithAfter(*c.After),
	).Paginate(s.db, &p2)
	s.assertIDs(p2, 1)
	s.assertBackwardOnly(c)

	var p3 []order
	_, c, _ = New(
		&cfg,
		WithBefore(*c.Before),
	).Paginate(s.db, &p3)
	s.assertIDs(p3, 2, 3)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateLimit() {
	s.givenOrders(10)

	var p1 []order
	_, c, _ := New(&Config{
		Limit: 1,
	}).Paginate(s.db, &p1)
	s.Len(p1, 1)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(&Config{
		Limit: 20,
		After: *c.After,
	}).Paginate(s.db, &p2)
	s.Len(p2, 9)
	s.assertBackwardOnly(c)

	var p3 []order
	_, c, _ = New(&Config{
		Limit:  100,
		Before: *c.Before,
	}).Paginate(s.db, &p3)
	s.Len(p3, 1)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateOrder() {
	s.givenOrders(20)

	var p1 []order
	_, c, _ := New(&Config{
		Order: ASC,
	}).Paginate(s.db, &p1)
	s.assertIDRange(p1, 1, 10)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(&Config{
		Before: *c.After,
		Order:  DESC,
	}).Paginate(s.db, &p2)
	s.assertIDRange(p2, 20, 11)
	s.assertForwardOnly(c)

	var p3 []order
	_, c, _ = New(&Config{
		Before: *c.After,
		Order:  ASC,
	}).Paginate(s.db, &p3)
	s.assertIDRange(p3, 1, 10)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateJoinQuery() {
	orders := s.givenOrders(3)
	// total 5 items
	s.givenItems(orders[0], 2) // ID: 1-3
	s.givenItems(orders[1], 2) // ID: 4-5
	s.givenItems(orders[2], 1) // ID: 6

	cfg := Config{
		Limit: 3,
	}

	stmt := s.db.
		Table("items").
		Joins("JOIN orders ON orders.id = items.order_id")

	var p1 []item
	_, c, _ := New(&cfg).Paginate(stmt, &p1)
	s.assertIDRange(p1, 5, 3)
	s.assertForwardOnly(c)

	var p2 []item
	_, c, _ = New(
		&cfg,
		WithAfter(*c.After),
	).Paginate(stmt, &p2)
	s.assertIDRange(p2, 2, 1)
	s.assertBackwardOnly(c)

	var p3 []item
	_, c, _ = New(
		&cfg,
		WithBefore(*c.Before),
	).Paginate(stmt, &p3)
	s.assertIDRange(p3, 5, 3)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateSpecialCharacter() {
	// ID ordered by Remark desc: 2, 1, 4, 3 (":" > "," > "&" > "%")
	s.givenOrders([]order{
		{Remark: ptrStr("a,b,c")},
		{Remark: ptrStr("a:b:c")},
		{Remark: ptrStr("a%b%c")},
		{Remark: ptrStr("a&b&c")},
	})

	cfg := Config{
		Keys:  []string{"Remark"},
		Limit: 3,
	}

	var p1 []order
	_, c, _ := New(&cfg).Paginate(s.db, &p1)
	s.assertIDs(p1, 2, 1, 4)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(
		&cfg,
		WithAfter(*c.After),
	).Paginate(s.db, &p2)
	s.assertIDs(p2, 3)
	s.assertBackwardOnly(c)

	var p3 []order
	_, c, _ = New(
		&cfg,
		WithBefore(*c.Before),
	).Paginate(s.db, &p3)
	s.assertIDs(p3, 2, 1, 4)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateConsistencyBetweenBuilderAndOptions() {
	s.givenOrders(10)

	var temp []order
	_, c, _ := New(&Config{
		Limit: 3,
	}).Paginate(s.db, &temp)

	anchorCursor := c.After

	var optOrders, builderOrders []order
	var optCursor, builderCursor cursor.Cursor

	// forward

	opts := []Option{
		WithKeys("Remark", "CreatedAt", "ID"),
		WithLimit(3),
		WithOrder(ASC),
		WithAfter(*anchorCursor),
	}
	_, optCursor, _ = New(opts...).Paginate(s.db, &optOrders)

	p := New()
	p.SetKeys("Remark", "CreatedAt", "ID")
	p.SetLimit(3)
	p.SetOrder(ASC)
	p.SetAfterCursor(*anchorCursor)
	_, builderCursor, _ = p.Paginate(s.db, &builderOrders)

	s.Equal(optOrders, builderOrders)
	s.Equal(optCursor, builderCursor)

	// backward

	opts = []Option{
		WithKeys("Remark", "CreatedAt", "ID"),
		WithLimit(3),
		WithOrder(ASC),
		WithBefore(*anchorCursor),
	}
	_, optCursor, _ = New(opts...).Paginate(s.db, &optOrders)

	p = New()
	p.SetKeys("Remark", "CreatedAt", "ID")
	p.SetLimit(3)
	p.SetOrder(ASC)
	p.SetBeforeCursor(*anchorCursor)
	_, builderCursor, _ = p.Paginate(s.db, &builderOrders)

	s.Equal(optOrders, builderOrders)
	s.Equal(optCursor, builderCursor)
}
