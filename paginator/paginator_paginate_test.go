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

/* data type */

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

func (s *paginatorSuite) TestPaginateSpecialCharacter() {
	// ordered by Remark desc -> 2, 1, 4, 3 (":" > "," > "&" > "%")
	s.givenOrders([]order{
		{ID: 1, Remark: ptrStr("a,b,c")},
		{ID: 2, Remark: ptrStr("a:b:c")},
		{ID: 3, Remark: ptrStr("a%b%c")},
		{ID: 4, Remark: ptrStr("a&b&c")},
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

/* cursor */

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

/* key */

func (s *paginatorSuite) TestPaginateSingleKey() {
	now := time.Now()
	// ordered by CreatedAt desc -> 1, 3, 2
	s.givenOrders([]order{
		{ID: 1, CreatedAt: now.Add(1 * time.Hour)},
		{ID: 2, CreatedAt: now.Add(-1 * time.Hour)},
		{ID: 3, CreatedAt: now},
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
	now := time.Now()
	// ordered by (CreatedAt desc, ID desc) -> 2, 3, 1
	s.givenOrders([]order{
		{ID: 1, CreatedAt: now},
		{ID: 2, CreatedAt: now.Add(1 * time.Hour)},
		{ID: 3, CreatedAt: now},
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

func (s *paginatorSuite) TestPaginateRulesShouldTakePrecedenceOverKeys() {
	now := time.Now()
	// ordered by ID desc -> 2, 1
	// ordered by CreatedAt desc -> 1, 2
	s.givenOrders([]order{
		{ID: 1, CreatedAt: now.Add(1 * time.Hour)},
		{ID: 2, CreatedAt: now},
	})

	cfg := Config{
		Rules: []Rule{
			{Key: "CreatedAt"},
		},
		Keys: []string{"ID"},
	}

	var orders []order
	_, _, _ = New(&cfg).Paginate(s.db, &orders)
	s.assertIDs(orders, 1, 2)
}

func (s *paginatorSuite) TestPaginateShouldUseGormColumnTag() {
	s.givenOrders([]order{
		{ID: 1, Remark: ptrStr("c")},
		{ID: 2, Remark: ptrStr("b")},
		{ID: 3, Remark: ptrStr("a")},
	})

	type order struct {
		ID          int
		Description string `json:"description" gorm:"type:varchar(255);column:remark"`
	}

	var orders []order
	result, _, _ := New(
		WithKeys("Description"),
	).Paginate(s.db, &orders)
	s.Nil(result.Error)
	s.assertIDs(orders, 1, 2, 3)
}

/* limit */

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

/* order */

func (s *paginatorSuite) TestPaginateOrder() {
	now := time.Now()
	// ordered by (CreatedAt desc, ID desc) -> 4, 2, 3, 1
	s.givenOrders([]order{
		{ID: 1, CreatedAt: now},
		{ID: 2, CreatedAt: now.Add(1 * time.Hour)},
		{ID: 3, CreatedAt: now},
		{ID: 4, CreatedAt: now.Add(2 * time.Hour)},
	})

	cfg := Config{
		Keys:  []string{"CreatedAt", "ID"},
		Limit: 2,
	}

	var p1 []order
	_, c, _ := New(
		&cfg,
		WithOrder(ASC),
	).Paginate(s.db, &p1)
	s.assertIDs(p1, 1, 3)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(
		&cfg,
		WithBefore(*c.After),
		WithOrder(DESC),
	).Paginate(s.db, &p2)
	s.assertIDs(p2, 4, 2)
	s.assertForwardOnly(c)

	var p3 []order
	_, c, _ = New(
		&cfg,
		WithBefore(*c.After),
		WithOrder(ASC),
	).Paginate(s.db, &p3)
	s.assertIDs(p3, 1, 3)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateOrderByKey() {
	now := time.Now()
	// ordered by (CreatedAt desc, ID asc) -> 4, 2, 1, 3
	s.givenOrders([]order{
		{ID: 1, CreatedAt: now},
		{ID: 2, CreatedAt: now.Add(1 * time.Hour)},
		{ID: 3, CreatedAt: now},
		{ID: 4, CreatedAt: now.Add(2 * time.Hour)},
	})

	cfg := Config{
		Rules: []Rule{
			{
				Key: "CreatedAt",
			},
			{
				Key:   "ID",
				Order: ASC,
			},
		},
		Limit: 2,
		Order: DESC, // default order for no order rule
	}

	var p1 []order
	_, c, _ := New(&cfg).Paginate(s.db, &p1)
	s.assertIDs(p1, 4, 2)
	s.assertForwardOnly(c)

	var p2 []order
	_, c, _ = New(
		&cfg,
		WithAfter(*c.After),
	).Paginate(s.db, &p2)
	s.assertIDs(p2, 1, 3)
	s.assertBackwardOnly(c)

	var p3 []order
	_, c, _ = New(
		&cfg,
		WithBefore(*c.Before),
	).Paginate(s.db, &p3)
	s.assertIDs(p3, 4, 2)
	s.assertForwardOnly(c)
}

/* join */

func (s *paginatorSuite) TestPaginateJoinQuery() {
	orders := s.givenOrders(3)
	// total 5 items
	// order 1 -> items (1, 2, 3)
	// order 2 -> items (4, 5)
	// order 3 -> items (6)
	s.givenItems(orders[0], 2)
	s.givenItems(orders[1], 2)
	s.givenItems(orders[2], 1)

	stmt := s.db.
		Table("items").
		Joins("JOIN orders ON orders.id = items.order_id")

	cfg := Config{
		Limit: 3,
	}

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

/* compatibility */

func (s *paginatorSuite) TestPaginateJoinQueryWithAlias() {
	orders := s.givenOrders(2)
	// total 6 items
	// order 1 -> items (1, 3, 5)
	// order 2 -> items (2, 4, 6)
	for i := 0; i < 3; i++ {
		s.givenItems(orders[0], 1)
		s.givenItems(orders[1], 1)
	}

	type itemDTO struct {
		ID      int
		OrderID int
	}

	stmt := s.db.
		Select("its.id AS id, ods.id AS order_id").
		Table("items AS its").
		Joins("JOIN orders AS ods ON ods.id = its.order_id")

	cfg := Config{
		Rules: []Rule{
			{
				Key:     "OrderID",
				SQLRepr: "ods.id",
			},
			{
				Key:     "ID",
				SQLRepr: "its.id",
			},
		},
		Limit: 3,
	}

	var p1 []itemDTO
	_, c, _ := New(&cfg).Paginate(stmt, &p1)
	s.assertIDs(p1, 6, 4, 2)
	s.assertForwardOnly(c)

	var p2 []itemDTO
	_, c, _ = New(
		&cfg,
		WithAfter(*c.After),
	).Paginate(stmt, &p2)
	s.assertIDs(p2, 5, 3, 1)
	s.assertBackwardOnly(c)

	var p3 []itemDTO
	_, c, _ = New(
		&cfg,
		WithBefore(*c.Before),
	).Paginate(stmt, &p3)
	s.assertIDs(p3, 6, 4, 2)
	s.assertForwardOnly(c)
}

func (s *paginatorSuite) TestPaginateConsistencyBetweenBuilderAndOptions() {
	s.givenOrders(10)

	var temp []order
	_, c, err := New(
		WithKeys("Remark", "CreatedAt", "ID"),
		WithLimit(3),
	).Paginate(s.db, &temp)
	if err != nil {
		s.FailNow(err.Error())
	}

	anchorCursor := *c.After

	var optOrders, builderOrders []order
	var optCursor, builderCursor cursor.Cursor

	// forward - keys

	opts := []Option{
		WithKeys("Remark", "CreatedAt", "ID"),
		WithLimit(3),
		WithOrder(ASC),
		WithAfter(anchorCursor),
	}
	_, optCursor, err = New(opts...).Paginate(s.db, &optOrders)
	s.Nil(err)

	p := New()
	p.SetKeys("Remark", "CreatedAt", "ID")
	p.SetLimit(3)
	p.SetOrder(ASC)
	p.SetAfterCursor(anchorCursor)
	_, builderCursor, err = p.Paginate(s.db, &builderOrders)
	s.Nil(err)

	s.Equal(optOrders, builderOrders)
	s.Equal(optCursor, builderCursor)

	// backward - keys

	opts = []Option{
		WithKeys("Remark", "CreatedAt", "ID"),
		WithLimit(3),
		WithOrder(ASC),
		WithBefore(anchorCursor),
	}
	_, optCursor, err = New(opts...).Paginate(s.db, &optOrders)
	s.Nil(err)

	p = New()
	p.SetKeys("Remark", "CreatedAt", "ID")
	p.SetLimit(3)
	p.SetOrder(ASC)
	p.SetBeforeCursor(anchorCursor)
	_, builderCursor, err = p.Paginate(s.db, &builderOrders)
	s.Nil(err)

	s.Equal(optOrders, builderOrders)
	s.Equal(optCursor, builderCursor)

	// forward - rules

	opts = []Option{
		WithRules([]Rule{
			{Key: "Remark"},
			{Key: "CreatedAt"},
			{Key: "ID"},
		}...),
		WithLimit(3),
		WithOrder(ASC),
		WithAfter(anchorCursor),
	}
	_, optCursor, err = New(opts...).Paginate(s.db, &optOrders)
	s.Nil(err)

	p = New()
	p.SetRules([]Rule{
		{Key: "Remark"},
		{Key: "CreatedAt"},
		{Key: "ID"},
	}...)
	p.SetLimit(3)
	p.SetOrder(ASC)
	p.SetAfterCursor(anchorCursor)
	_, builderCursor, err = p.Paginate(s.db, &builderOrders)
	s.Nil(err)

	s.Equal(optOrders, builderOrders)
	s.Equal(optCursor, builderCursor)

	// backward - rules

	opts = []Option{
		WithRules([]Rule{
			{Key: "Remark"},
			{Key: "CreatedAt"},
			{Key: "ID"},
		}...),
		WithLimit(3),
		WithOrder(ASC),
		WithBefore(anchorCursor),
	}
	_, optCursor, err = New(opts...).Paginate(s.db, &optOrders)
	s.Nil(err)

	p = New()
	p.SetRules([]Rule{
		{Key: "Remark"},
		{Key: "CreatedAt"},
		{Key: "ID"},
	}...)
	p.SetLimit(3)
	p.SetOrder(ASC)
	p.SetBeforeCursor(anchorCursor)
	_, builderCursor, err = p.Paginate(s.db, &builderOrders)
	s.Nil(err)

	s.Equal(optOrders, builderOrders)
	s.Equal(optCursor, builderCursor)
}
