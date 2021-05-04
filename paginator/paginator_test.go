package paginator

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/cursor"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPaginator(t *testing.T) {
	suite.Run(t, &paginatorSuite{})
}

/* models */

type order struct {
	ID        int       `gorm:"primaryKey"`
	Remark    *string   `gorm:"varchar(30)"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
}

type item struct {
	ID      int     `gorm:"primaryKey"`
	Name    string  `gorm:"type:varchar(30);not null"`
	Remark  *string `gorm:"type:varchar(30)"`
	OrderID int     `gorm:"type:integer;not null"`
	Order   Order   `gorm:"foreignKey:OrderID"`
}

/* paginator suite */

type paginatorSuite struct {
	suite.Suite
	db *gorm.DB
}

/* setup */

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

/* teardown */

func (s *paginatorSuite) TearDownTest() {
	s.db.Exec("TRUNCATE orders, items RESTART IDENTITY;")
}

func (s *paginatorSuite) TearDownSuite() {
	s.db.Migrator().DropTable(&item{}, &order{})
}

/* fixtures */

func (s *paginatorSuite) givenOrders(numOrOrders interface{}) (orders []order) {
	switch v := numOrOrders.(type) {
	case int:
		for i := 0; i < v; i++ {
			orders = append(orders, order{
				CreatedAt: time.Now().Add(time.Duration(i) * time.Hour),
			})
		}
	case []order:
		orders = v
	default:
		panic("givenOrders: numOrOrders should be number or orders")
	}
	for i := 0; i < len(orders); i++ {
		if err := s.db.Create(&orders[i]).Error; err != nil {
			panic(err.Error())
		}
	}
	return
}

func (s *paginatorSuite) givenItems(order order, numOrItems interface{}) (items []item) {
	switch v := numOrItems.(type) {
	case int:
		for i := 0; i < v; i++ {
			items = append(items, item{
				Name:    fmt.Sprintf("item %d", i+1),
				OrderID: order.ID,
			})
		}
	case []item:
		items = v
	default:
		panic("givenItems: numOrItems should be number or items")
	}
	for i := 0; i < len(items); i++ {
		if err := s.db.Create(&items[i]).Error; err != nil {
			panic(err.Error())
		}
	}
	return
}

/* assertions */

func (s *paginatorSuite) assertIDRange(result interface{}, fromID, toID int) {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Slice {
		panic("assertIDRange: result should be a slice")
	}
	s.Equal(
		int(math.Abs(float64(fromID-toID))+1),
		rv.Len(),
	)
	cur, vector := fromID, 1
	if fromID > toID {
		vector = -1
	}
	for i := 0; i < rv.Len(); i++ {
		e := rv.Index(i)
		if e.Kind() == reflect.Ptr {
			e = e.Elem()
		}
		s.Equal(cur, e.FieldByName("ID").Interface())
		cur += vector
	}
}

func (s *paginatorSuite) assertIDs(result interface{}, ids ...int) {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Slice {
		panic("assertIDs: result should be a slice")
	}
	s.Equal(len(ids), rv.Len())

	for i := 0; i < rv.Len(); i++ {
		e := rv.Index(i)
		if e.Kind() == reflect.Ptr {
			e = e.Elem()
		}
		s.Equal(ids[i], e.FieldByName("ID").Interface())
	}
}

func (s *paginatorSuite) assertForwardOnly(c cursor.Cursor) {
	s.NotNil(c.After)
	s.Nil(c.Before)
}

func (s *paginatorSuite) assertBackwardOnly(c cursor.Cursor) {
	s.Nil(c.After)
	s.NotNil(c.Before)
}

func (s *paginatorSuite) assertBothDirections(c cursor.Cursor) {
	s.NotNil(c.After)
	s.NotNil(c.Before)
}

func (s *paginatorSuite) assertNoMore(c cursor.Cursor) {
	s.Nil(c.After)
	s.Nil(c.Before)
}

/* util */

func ptrStr(v string) *string {
	return &v
}
