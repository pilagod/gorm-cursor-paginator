package paginator

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPaginator(t *testing.T) {
	suite.Run(t, &paginatorSuite{})
}

/* test model */

type order struct {
	ID        int       `gorm:"primaryKey"`
	Name      *string   `gorm:"type:varchar(30)"`
	Items     []item    `gorm:"foreignKey:OrderID"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
}

type item struct {
	ID      int   `gorm:"primaryKey"`
	OrderID int   `gorm:"type:integer;not null"`
	Order   Order `gorm:"foreignKey:OrderID"`
}

/* suite */

type paginatorSuite struct {
	suite.Suite
	db *gorm.DB
}

/* suite setup */

func (s *paginatorSuite) SetupSuite() {
	db, err := gorm.Open(
		postgres.Open("host=localhost port=8765 dbname=test user=test password=test sslmode=disable"),
		&gorm.Config{},
	)
	if err != nil {
		s.FailNow(err.Error())
	}
	s.db = db
	s.db.AutoMigrate(&order{}, &item{})
}

func (s *paginatorSuite) TearDownTest() {
	s.db.Exec("TRUNCATE orders, items RESTART IDENTITY;")
}

func (s *paginatorSuite) TearDownSuite() {
	s.db.Migrator().DropTable(&item{}, &order{})
}

/* suite test cases */

func (s *paginatorSuite) TestPaginateDefaultOptions() {
	var orders = s.givenOrders(12)

	var o1 []order
	cursor := s.paginate(s.db, &o1, pq{})
	s.assertOrders(orders, 11, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	cursor = s.paginate(s.db, &o2, pq{
		After: cursor.After,
	})
	s.assertOrders(orders, 1, 0, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	cursor = s.paginate(s.db, &o3, pq{
		Before: cursor.Before,
	})
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateDefaultOptionsForSliceStructPointers() {
	var ptrOrders = s.givenPtrOrders(12)

	var o1 []*order
	cursor := s.paginate(s.db, &o1, pq{})
	s.assertPtrOrders(ptrOrders, 11, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []*order
	cursor = s.paginate(s.db, &o2, pq{
		After: cursor.After,
	})
	s.assertPtrOrders(ptrOrders, 1, 0, o2)
	s.assertOnlyBefore(cursor)

	var o3 []*order
	cursor = s.paginate(s.db, &o3, pq{
		Before: cursor.Before,
	})
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateAfterCursorShouldTakePrecedenceOverBeforeCursor() {
	var orders = s.givenOrders(10)

	var o1 []order
	cursor := s.paginate(s.db, &o1, pq{
		Limit: pqLimit(3),
	})
	s.assertOnlyAfter(cursor)

	var o2 []order
	cursor = s.paginate(s.db, &o2, pq{
		After: cursor.After,
		Limit: pqLimit(3),
	})
	s.assertBoth(cursor)

	var o3 []order
	cursor = s.paginate(s.db, &o3, pq{
		After:  cursor.After,
		Before: cursor.Before,
	})
	s.assertOrders(orders, 3, 0, o3)
	s.assertOnlyBefore(cursor)
}

func (s *paginatorSuite) TestPaginateSingleKey() {
	var orders = s.givenCustomOrders([]order{
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now().Add(-1 * time.Hour)},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now().Add(-2 * time.Hour)},
	})
	var keys = []string{"CreatedAt"}

	var o1 []order
	cursor := s.paginate(s.db, &o1, pq{
		Keys:  keys,
		Limit: pqLimit(2),
	})
	s.assertOrders(orders, 3, 1, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	cursor = s.paginate(s.db, &o2, pq{
		Keys:  keys,
		After: cursor.After,
	})
	s.assertOrders(orders, 0, 4, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	cursor = s.paginate(s.db, &o3, pq{
		Keys:   keys,
		Before: cursor.Before,
	})
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateMultipleKeys() {
	var orders = s.givenCustomOrders([]order{
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(3 * time.Hour)},
		{CreatedAt: time.Now().Add(1 * time.Hour)},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
	})
	var keys = []string{"CreatedAt", "ID"}

	var o1 []order
	cursor := s.paginate(s.db, &o1, pq{
		Keys:  keys,
		Limit: pqLimit(3),
	})
	s.assertOrders(orders, 2, 0, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	cursor = s.paginate(s.db, &o2, pq{
		Keys:  keys,
		After: cursor.After,
	})
	s.assertOrders(orders, 3, 1, o2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	cursor = s.paginate(s.db, &o3, pq{
		Keys:   keys,
		Before: cursor.Before,
	})
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateLimitOption() {
	s.givenOrders(5)

	var o1 []order
	cursor := s.paginate(s.db, &o1, pq{
		Limit: pqLimit(1),
	})
	s.Len(o1, 1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	cursor = s.paginate(s.db, &o2, pq{
		After: cursor.After,
		Limit: pqLimit(20),
	})
	s.Len(o2, 4)
	s.assertOnlyBefore(cursor)

	var o3 []order
	cursor = s.paginate(s.db, &o3, pq{
		Before: cursor.Before,
	})
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateOrderOption() {
	var orders = s.givenOrders(10)

	var o1 []order
	cursor := s.paginate(s.db, &o1, pq{
		Limit: pqLimit(3),
		Order: pqOrder(ASC),
	})
	s.assertOrders(orders, 0, 2, o1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	cursor = s.paginate(s.db, &o2, pq{
		Before: cursor.After,
		Limit:  pqLimit(3),
	})
	s.assertOrders(orders, 5, 3, o2)
	s.assertBoth(cursor)

	var o3 []order
	cursor = s.paginate(s.db, &o3, pq{
		Before: cursor.After,
		Order:  pqOrder(ASC),
	})
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
	cursor := s.paginate(stmt, &i1, pq{
		Limit: pqLimit(3),
	})
	s.Len(i1, 3)
	s.assertItems(items, 4, 2, i1)
	s.assertOnlyAfter(cursor)

	var i2 []item
	cursor = s.paginate(stmt, &i2, pq{
		After: cursor.After,
	})
	s.Len(i2, 2)
	s.assertItems(items, 1, 0, i2)
	s.assertOnlyBefore(cursor)

	var i3 []item
	cursor = s.paginate(stmt, &i3, pq{
		Before: cursor.Before,
	})
	s.Equal(i1, i3)
	s.assertOnlyAfter(cursor)
}

func (s *paginatorSuite) TestPaginateSpecialCharacter() {
	s.givenCustomOrders([]order{
		{Name: pqString("a,b,c")},
		{Name: pqString("a:b:c")},
		{Name: pqString("a%b%c")},
	})
	var keys = []string{"Name"}

	var o1 []order
	cursor := s.paginate(s.db, &o1, pq{
		Keys:  keys,
		Limit: pqLimit(1),
	})
	s.Len(o1, 1)
	s.assertOnlyAfter(cursor)

	var o2 []order
	cursor = s.paginate(s.db, &o2, pq{
		Keys:  keys,
		After: cursor.After,
	})
	s.Len(o2, 2)
	s.assertOnlyBefore(cursor)

	var o3 []order
	cursor = s.paginate(s.db, &o3, pq{
		Keys:   keys,
		Before: cursor.Before,
	})
	s.Len(o3, 1)
	s.Equal(o1, o3)
	s.assertOnlyAfter(cursor)
}

/* util */

func (s *paginatorSuite) paginate(stmt *gorm.DB, out interface{}, q pq) Cursor {
	p := q.Paginator()
	if err := p.Paginate(stmt, out).Error; err != nil {
		s.FailNow(err.Error())
	}
	return p.GetNextCursor()
}

// pq stands for paging query
type pq struct {
	Keys   []string
	After  *string
	Before *string
	Limit  *int
	Order  *Order
}

func (q pq) Paginator() *Paginator {
	p := New()
	if q.Keys != nil {
		p.SetKeys(q.Keys...)
	}
	if q.After != nil {
		p.SetAfterCursor(*q.After)
	}
	if q.Before != nil {
		p.SetBeforeCursor(*q.Before)
	}
	if q.Limit != nil {
		p.SetLimit(*q.Limit)
	}
	if q.Order != nil {
		p.SetOrder(*q.Order)
	}
	return p
}

func pqString(str string) *string {
	return &str
}

func pqLimit(limit int) *int {
	return &limit
}

func pqOrder(order Order) *Order {
	return &order
}

/* order */

func (s *paginatorSuite) givenOrders(n int) []order {
	orders := make([]order, n)
	for i := 0; i < n; i++ {
		orders[i] = order{}
	}
	return s.createOrders(orders)
}

func (s *paginatorSuite) givenPtrOrders(n int) []*order {
	var result []*order
	orders := s.givenOrders(n)
	for i := 0; i < len(orders); i++ {
		result = append(result, &orders[i])
	}
	return result
}

func (s *paginatorSuite) givenCustomOrders(orders []order) []order {
	return s.createOrders(orders)
}

func (s *paginatorSuite) createOrders(orders []order) []order {
	for i := 0; i < len(orders); i++ {
		if err := s.db.Create(&orders[i]).Error; err != nil {
			s.FailNow(err.Error())
		}
	}
	return orders
}

/* item */

func (s *paginatorSuite) givenItems(orderID int, n int) []item {
	items := make([]item, n)
	for i := 0; i < n; i++ {
		items[i] = item{
			OrderID: orderID,
		}
	}
	return s.createItems(items)
}

func (s *paginatorSuite) createItems(items []item) []item {
	for i := 0; i < len(items); i++ {
		if err := s.db.Create(&items[i]).Error; err != nil {
			s.FailNow(err.Error())
		}
	}
	return items
}

/* assert */

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
