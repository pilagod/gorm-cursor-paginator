package paginator

import (
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

func TestPaginator(t *testing.T) {
	suite.Run(t, &paginatorSuite{})
}

type paginatorSuite struct {
	baseSuite
}

/* test cases */

func (s *paginatorSuite) TestPaginateDefaultOptions() {
	var orders = s.givenOrders(12)

	var o1 []order
	p1 := newPaginator(pq{})
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 11, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := newPaginator(pq{After: cursor.After})
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 1, 0, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := newPaginator(pq{Before: cursor.Before})
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateDefaultOptionsForSliceStructPointers() {
	var ptrOrders = s.givenPtrOrders(12)

	var o1 []*order
	p1 := newPaginator(pq{})
	cursor := s.paginate(p1, s.db, &o1)
	s.assertPtrOrders(ptrOrders, 11, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []*order
	p2 := newPaginator(pq{After: cursor.After})
	cursor = s.paginate(p2, s.db, &o2)
	s.assertPtrOrders(ptrOrders, 1, 0, o2)
	s.assertOnlyBefore(cursor)

	var o3 []*order
	p3 := newPaginator(pq{Before: cursor.Before})
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateAfterCursorShouldTakePrecedenceOverBeforeCursor() {
	var orders = s.givenOrders(10)

	var o1 []order
	p1 := newPaginator(pq{Limit: pqLimit(3)})
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := newPaginator(pq{
		After: cursor.After,
		Limit: pqLimit(3),
	})
	p2.SetAfterCursor(*cursor.After)
	cursor = s.paginate(p2, s.db, &o2)
	s.assertBoth(cursor)

	var o3 []order
	p3 := newPaginator(pq{
		After:  cursor.After,
		Before: cursor.Before,
	})
	cursor = s.paginate(p3, s.db, &o3)
	s.assertOrders(orders, 3, 0, o3)
	s.assertOnlyBefore(cursor)
}

func (s *paginatorSuite) TestPaginateSingleKey() {
	var p = func(q pq) *Paginator {
		p := newPaginator(q)
		p.SetKeys("CreatedAt")
		return p
	}
	var orders = s.givenCustomOrders([]order{
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now().Add(-1 * time.Hour)},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now().Add(-2 * time.Hour)},
	})

	var o1 []order
	p1 := p(pq{Limit: pqLimit(2)})
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 3, 1, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := p(pq{After: cursor.After})
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 0, 4, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := p(pq{Before: cursor.Before})
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateMultipleKeys() {
	var p = func(q pq) *Paginator {
		p := newPaginator(q)
		p.SetKeys("CreatedAt", "ID")
		return p
	}
	var orders = s.givenCustomOrders([]order{
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(3 * time.Hour)},
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
	})

	var o1 []order
	p1 := p(pq{Limit: pqLimit(3)})
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 2, 0, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := p(pq{After: cursor.After})
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 3, 1, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := p(pq{Before: cursor.Before})
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateLimitOption() {
	s.givenOrders(5)

	var o1 []order
	p1 := newPaginator(pq{Limit: pqLimit(1)})
	cursor := s.paginate(p1, s.db, &o1)
	s.Len(o1, 1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := newPaginator(pq{
		After: cursor.After,
		Limit: pqLimit(20),
	})
	cursor = s.paginate(p2, s.db, &o2)
	s.Len(o2, 4)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := newPaginator(pq{Before: cursor.Before})
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateOrderOption() {
	var orders = s.givenOrders(10)

	var o1 []order
	p1 := newPaginator(pq{
		Limit: pqLimit(3),
		Order: pqOrder(ASC),
	})
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 0, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := newPaginator(pq{
		Before: cursor.After,
		Limit:  pqLimit(3),
	})
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 5, 3, o2)
	s.assertBoth(cursor)

	var o3 []order
	p3 := newPaginator(pq{
		Before: cursor.After,
		Order:  pqOrder(ASC),
	})
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateJoinQuery() {
	var orders = s.givenOrders(3)
	var items = s.givenItems(orders[0].ID, 5)
	s.givenItems(orders[1].ID, 2)
	s.givenItems(orders[2].ID, 1)

	var stmt = s.db.
		Table("items").
		Joins("LEFT JOIN orders ON orders.id = items.order_id").
		Where("orders.id = ?", orders[0].ID)

	var i1 []item
	p1 := newPaginator(pq{Limit: pqLimit(3)})
	cursor := s.paginate(p1, stmt, &i1)
	s.Len(i1, 3)
	s.assertItems(items, 4, 2, i1)
	s.assertOnlyAfter(cursor)

	var i2 []item
	p2 := newPaginator(pq{After: cursor.After})
	cursor = s.paginate(p2, stmt, &i2)
	s.Len(i2, 2)
	s.assertItems(items, 1, 0, i2)
	s.assertOnlyBefore(cursor)

	var i3 []item
	p3 := newPaginator(pq{Before: cursor.Before})
	cursor = s.paginate(p3, stmt, &i3)
	s.Equal(i1, i3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateSpecialCharacter() {
	var p = func(q pq) *Paginator {
		p := newPaginator(q)
		p.SetKeys("Name")
		return p
	}
	s.givenCustomOrders([]order{
		{Name: pqString("a,b,c")},
		{Name: pqString("a:b:c")},
		{Name: pqString("a%b%c")},
	})

	var o1 []order
	p1 := p(pq{Limit: pqLimit(1)})
	cursor := s.paginate(p1, s.db, &o1)
	s.Len(o1, 1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := p(pq{After: cursor.After})
	cursor = s.paginate(p2, s.db, &o2)
	s.Len(o2, 2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := p(pq{Before: cursor.Before})
	cursor = s.paginate(p3, s.db, &o3)
	s.Len(o3, 1)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}
