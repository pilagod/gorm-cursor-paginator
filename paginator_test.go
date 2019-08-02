package paginator

import (
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

func TestPaginator(t *testing.T) {
	suite.Run(t, &paginatorSuite{})
}

type paginatorSuite struct {
	suite.Suite
	db *gorm.DB
}

type order struct {
	ID        int       `gorm:"primary_key"`
	Items     []item    `gorm:"foreignkey:OrderID"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
}

type item struct {
	ID      int   `gorm:"primary_key"`
	OrderID int   `gorm:"type:integer;not null"`
	Order   Order `gorm:"foreignkey:OrderID"`
}

func (s *paginatorSuite) SetupSuite() {
	db, err := gorm.Open("postgres", "host=localhost port=8765 dbname=test user=test password=test sslmode=disable")
	if err != nil {
		s.FailNow(err.Error())
	}
	s.db = db
	s.db.AutoMigrate(&order{}, &item{})
	s.db.Model(&item{}).AddForeignKey("order_id", "orders(id)", "CASCADE", "CASCADE")
}

func (s *paginatorSuite) TearDownTest() {
	s.db.Exec("TRUNCATE orders, items RESTART IDENTITY;")
}

func (s *paginatorSuite) TearDownSuite() {
	s.db.DropTable(&item{}, &order{})
	s.db.Close()
}

func (s *paginatorSuite) TestPaginateWithDefaultOptions() {
	var orders = s.givenOrders(12)

	var o1 []order
	p1 := New()
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 11, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := New()
	p2.SetAfterCursor(*cursor.After)
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 1, 0, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := New()
	p3.SetBeforeCursor(*cursor.Before)
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateWithSliceStructPointersDefaultOptions() {
	var ordersPtrs []*order
	var orders = s.givenOrders(12)

	// Create Slice of order pointers
	for i := 0; i < len(orders); i++ {
		ordersPtrs = append(ordersPtrs, &orders[i])
	}

	var o1 []*order
	p1 := New()
	cursor := s.paginate(p1, s.db, &o1)
	s.assertPtrOrders(ordersPtrs, 11, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []*order
	p2 := New()
	p2.SetAfterCursor(*cursor.After)
	cursor = s.paginate(p2, s.db, &o2)
	s.assertPtrOrders(ordersPtrs, 1, 0, o2)
	s.assertOnlyBefore(cursor)

	var o3 []*order
	p3 := New()
	p3.SetBeforeCursor(*cursor.Before)
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateAfterCursorShouldTakePrecedenceOverBeforeCursor() {
	var orders = s.givenOrders(10)

	var o1 []order
	p1 := New()
	p1.SetLimit(3)
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := New()
	p2.SetLimit(3)
	p2.SetAfterCursor(*cursor.After)
	cursor = s.paginate(p2, s.db, &o2)
	s.assertBoth(cursor)

	var o3 []order
	p3 := New()
	p3.SetAfterCursor(*cursor.After)
	p3.SetBeforeCursor(*cursor.Before)
	cursor = s.paginate(p3, s.db, &o3)
	s.assertOrders(orders, 3, 0, o3)
	s.assertOnlyBefore(cursor)
}

func (s *paginatorSuite) TestPaginateWithSingleKey() {
	var orders = []order{
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now().Add(-1 * time.Hour)},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now().Add(-2 * time.Hour)},
	}
	s.createOrders(orders)
	var p = func() *Paginator {
		p := New()
		p.SetKeys("CreatedAt")
		return p
	}
	var o1 []order
	p1 := p()
	p1.SetLimit(2)
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 3, 1, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := p()
	p2.SetAfterCursor(*cursor.After)
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 0, 4, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := New()
	p3.SetKeys("CreatedAt")
	p3.SetBeforeCursor(*cursor.Before)
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateWithMultipleKeys() {
	var orders = []order{
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(3 * time.Hour)},
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
	}
	s.createOrders(orders)
	var p = func() *Paginator {
		p := New()
		p.SetKeys("CreatedAt", "ID")
		return p
	}
	var o1 []order
	p1 := p()
	p1.SetLimit(3)
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 2, 0, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := p()
	p2.SetAfterCursor(*cursor.After)
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 3, 1, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := p()
	p3.SetBeforeCursor(*cursor.Before)
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateWithLimitOption() {
	s.givenOrders(5)

	var o1 []order
	p1 := New()
	p1.SetLimit(1)
	cursor := s.paginate(p1, s.db, &o1)
	s.Len(o1, 1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := New()
	p2.SetAfterCursor(*cursor.After)
	p2.SetLimit(20)
	cursor = s.paginate(p2, s.db, &o2)
	s.Len(o2, 4)
	s.assertOnlyBefore(cursor)

	var o3 []order
	p3 := New()
	p3.SetBeforeCursor(*cursor.Before)
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateWithOrderOption() {
	var orders = s.givenOrders(10)

	var o1 []order
	p1 := New()
	p1.SetLimit(3)
	p1.SetOrder(ASC)
	cursor := s.paginate(p1, s.db, &o1)
	s.assertOrders(orders, 0, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	p2 := New()
	p2.SetLimit(3)
	p2.SetBeforeCursor(*cursor.After)
	cursor = s.paginate(p2, s.db, &o2)
	s.assertOrders(orders, 5, 3, o2)
	s.assertBoth(cursor)

	var o3 []order
	p3 := New()
	p3.SetBeforeCursor(*cursor.After)
	p3.SetOrder(ASC)
	cursor = s.paginate(p3, s.db, &o3)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateWithJoinQuery() {
	var orders = s.givenOrders(3)
	var items = s.givenItems(orders[0].ID, 5)
	s.givenItems(orders[1].ID, 2)
	s.givenItems(orders[2].ID, 1)

	var stmt = s.db.
		Table("items").
		Joins("LEFT JOIN orders ON orders.id = items.order_id").
		Where("orders.id = ?", orders[0].ID)

	var i1 []item
	p1 := New()
	p1.SetLimit(3)
	cursor := s.paginate(p1, stmt, &i1)
	s.Len(i1, 3)
	s.assertItems(items, 4, 2, i1)
	s.assertOnlyAfter(cursor)

	var i2 []item
	p2 := New()
	p2.SetAfterCursor(*cursor.After)
	cursor = s.paginate(p2, stmt, &i2)
	s.Len(i2, 2)
	s.assertItems(items, 1, 0, i2)
	s.assertOnlyBefore(cursor)

	var i3 []item
	p3 := New()
	p3.SetBeforeCursor(*cursor.Before)
	cursor = s.paginate(p3, stmt, &i3)
	s.Equal(i1, i3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) paginate(p *Paginator, stmt *gorm.DB, out interface{}) Cursor {
	if err := p.Paginate(stmt, out).Error; err != nil {
		s.FailNow(err.Error())
	}
	return p.GetNextCursor()
}

func (s *paginatorSuite) givenOrders(n int) []order {
	orders := make([]order, n)
	for i := 0; i < n; i++ {
		orders[i] = order{}
	}
	s.createOrders(orders)
	return orders
}

func (s *paginatorSuite) createOrders(orders []order) {
	for i := 0; i < len(orders); i++ {
		if err := s.db.Create(&orders[i]).Error; err != nil {
			s.FailNow(err.Error())
		}
	}
}

func (s *paginatorSuite) givenItems(orderID int, n int) []item {
	items := make([]item, n)
	for i := 0; i < n; i++ {
		items[i] = item{
			OrderID: orderID,
		}
	}
	s.createItems(items)
	return items
}

func (s *paginatorSuite) createItems(items []item) {
	for i := 0; i < len(items); i++ {
		if err := s.db.Create(&items[i]).Error; err != nil {
			s.FailNow(err.Error())
		}
	}
}

func (s *paginatorSuite) assertOnlyAfter(cursor Cursor) {
	s.NotNil(cursor.After)
	s.Nil(cursor.Before)
}

func (s *paginatorSuite) assertOnlyBefore(cursor Cursor) {
	s.Nil(cursor.After)
	s.NotNil(cursor.Before)
}

func (s *paginatorSuite) assertBoth(cursor Cursor) {
	s.NotNil(cursor.After)
	s.NotNil(cursor.Before)
}

func (s *paginatorSuite) assertOrders(expected []order, head, tail int, got []order) {
	s.Equal(expected[head].ID, got[first(got)].ID)
	s.Equal(expected[tail].ID, got[last(got)].ID)
}

func (s *paginatorSuite) assertPtrOrders(expected []*order, head, tail int, got []*order) {
	s.Equal(expected[head].ID, got[first(got)].ID)
	s.Equal(expected[tail].ID, got[last(got)].ID)
}

func (s *paginatorSuite) assertItems(expected []item, head, tail int, got []item) {
	s.Equal(expected[head].ID, got[first(got)].ID)
	s.Equal(expected[tail].ID, got[last(got)].ID)
}

func first(values interface{}) int {
	return 0
}

func last(values interface{}) int {
	return reflect.ValueOf(values).Len() - 1
}
