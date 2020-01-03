package paginator

import (
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

/* test model */

type order struct {
	ID        int       `gorm:"primary_key"`
	Name      *string   `gorm:"type:varchar(30)"`
	Items     []item    `gorm:"foreignkey:OrderID"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
}

type item struct {
	ID      int   `gorm:"primary_key"`
	OrderID int   `gorm:"type:integer;not null"`
	Order   Order `gorm:"foreignkey:OrderID"`
}

/* suite */

type baseSuite struct {
	suite.Suite
	db *gorm.DB
}

/* suite setup */

func (s *baseSuite) SetupSuite() {
	db, err := gorm.Open("postgres", "host=localhost port=8765 dbname=test user=test password=test sslmode=disable")
	if err != nil {
		s.FailNow(err.Error())
	}
	s.db = db
	s.db.AutoMigrate(&order{}, &item{})
	s.db.Model(&item{}).AddForeignKey("order_id", "orders(id)", "CASCADE", "CASCADE")
}

func (s *baseSuite) TearDownTest() {
	s.db.Exec("TRUNCATE orders, items RESTART IDENTITY;")
}

func (s *baseSuite) TearDownSuite() {
	s.db.DropTable(&item{}, &order{})
	s.db.Close()
}

/* order */

func (s *baseSuite) givenOrders(n int) []order {
	orders := make([]order, n)
	for i := 0; i < n; i++ {
		orders[i] = order{}
	}
	return s.givenCustomOrders(orders)
}

func (s *baseSuite) givenCustomOrders(orders []order) []order {
	s.createOrders(orders)
	return orders
}

func (s *baseSuite) createOrders(orders []order) {
	for i := 0; i < len(orders); i++ {
		if err := s.db.Create(&orders[i]).Error; err != nil {
			s.FailNow(err.Error())
		}
	}
}

func (s *baseSuite) givenPtrOrders(n int) []*order {
	var result []*order
	orders := s.givenOrders(n)
	for i := 0; i < len(orders); i++ {
		result = append(result, &orders[i])
	}
	return result
}

/* item */

func (s *baseSuite) givenItems(orderID int, n int) []item {
	items := make([]item, n)
	for i := 0; i < n; i++ {
		items[i] = item{
			OrderID: orderID,
		}
	}
	s.createItems(items)
	return items
}

func (s *baseSuite) createItems(items []item) {
	for i := 0; i < len(items); i++ {
		if err := s.db.Create(&items[i]).Error; err != nil {
			s.FailNow(err.Error())
		}
	}
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
