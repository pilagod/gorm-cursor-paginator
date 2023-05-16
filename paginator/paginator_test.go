package paginator

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

func TestPaginator(t *testing.T) {
	suite.Run(t, &paginatorSuite{})
}

/* models */

type order struct {
	ID        int       `gorm:"primaryKey"`
	Remark    *string   `gorm:"type:varchar(30)"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
	Data      JSON      `gorm:"type:jsonb"`
}

type item struct {
	ID      int     `gorm:"primaryKey"`
	Name    string  `gorm:"type:varchar(30);not null"`
	Remark  *string `gorm:"type:varchar(30)"`
	OrderID int     `gorm:"type:integer;not null"`
	Order   Order   `gorm:"foreignKey:OrderID"`
}

/* JSON type taken from https://gorm.io/docs/data_types.html */

type JSON struct {
	json.RawMessage
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON{result}
	return err
}

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	return j.MarshalJSON()
}

func (j JSON) GetCustomTypeValue(meta interface{}) (interface{}, error) {
	// The decoding & mapping logic should be implemented by custom type provider
	d := map[string]interface{}{}
	err := json.Unmarshal(j.RawMessage, &d)
	if err != nil {
		return nil, errors.New("json unmarshalling failed")
	}

	i := d[meta.(string)]

	// remove quotes from string for SQL comparisons
	if s, ok := i.(string); ok {
		return strings.ReplaceAll(s, "\"", ""), nil
	}

	return i, nil
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
			// prepare JSON data
			data, err := json.Marshal(map[string]interface{}{
				"keyInt":    i,
				"keyString": fmt.Sprintf("%d", i),
			})
			if err != nil {
				panic(err.Error())
			}
			j := JSON{}
			if err := j.Scan(data); err != nil {
				panic(err.Error())
			}

			orders = append(orders, order{
				CreatedAt: time.Now().Add(time.Duration(i) * time.Hour),
				Data:      j,
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

func (s *paginatorSuite) assertForwardOnly(c Cursor) {
	s.NotNil(c.After)
	s.Nil(c.Before)
}

func (s *paginatorSuite) assertBackwardOnly(c Cursor) {
	s.Nil(c.After)
	s.NotNil(c.Before)
}

func (s *paginatorSuite) assertBothDirections(c Cursor) {
	s.NotNil(c.After)
	s.NotNil(c.Before)
}

func (s *paginatorSuite) assertNoMore(c Cursor) {
	s.Nil(c.After)
	s.Nil(c.Before)
}

/* util */

func ptrStr(v string) *string {
	return &v
}
