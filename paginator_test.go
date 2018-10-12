package paginator

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func assertEqual(t *testing.T, caseInfo interface{}, got, expected interface{}) {
	if got != expected {
		t.Errorf("[case %v] expected `%v`, but got `%v`", caseInfo, expected, got)
	}
}

func TestInitOptions(t *testing.T) {
	p := New()

	p.initOptions()

	if len(p.keys) != 1 || p.keys[0] != "ID" {
		t.Errorf("paginator should have only one default key: ID, but got %v", p.keys)
	}
	if p.limit != 10 {
		t.Errorf("paginator should have default limit: 10, but got %v", p.limit)
	}
	if p.order != DESC {
		t.Errorf("paginator should have default order: DESC, but got %v", p.order)
	}
}

func mockDB() (sqlmock.Sqlmock, *gorm.DB) {
	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatalf("can not create sqlmock: %s", err)
	}
	gormDB, gerr := gorm.Open("postgres", db)

	if gerr != nil {
		log.Fatalf("can not open gorm connection: %s", err)
	}
	gormDB.LogMode(true)

	return mock, gormDB
}

type Dummy struct {
	A string
	B string
}

func getDummyRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{})
}

func TestPaginateWithDefaultOptions(t *testing.T) {
	mock, db := mockDB()

	mock.ExpectQuery(
		`.*ORDER BY id DESC LIMIT 11`,
	).WillReturnRows(getDummyRows())

	var dummy Dummy

	p := New()

	p.Paginate(db, &dummy)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expections: %s", err)
	}
}

func TestPaginateWithLimit(t *testing.T) {
	mock, db := mockDB()

	for _, limit := range []int{1, 2, 3, 5, 8, 13, 21, 34} {
		mock.ExpectQuery(
			fmt.Sprintf(`.*LIMIT %d`, limit+1),
		).WillReturnRows(getDummyRows())

		var dummy Dummy

		p := New()

		p.SetLimit(limit)
		p.Paginate(db, &dummy)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expections: %s", err)
	}
}

func newDummyPaginatorWithAfterCursor() Paginator {
	p := New()
	p.SetKeys("A", "B")
	p.SetAfterCursor(p.encode(reflect.ValueOf(Dummy{"A", "B"})))
	return p
}

func TestPaginateWithAfterCursor(t *testing.T) {
	mock, db := mockDB()

	for _, testCase := range []struct {
		order            order
		expectedOperator string
		expectedOrder    order
	}{
		{ASC, ">", ASC},
		{DESC, "<", DESC},
	} {
		mock.ExpectQuery(
			fmt.Sprintf(
				`.*WHERE.*\(a %[1]v \$1 OR a = \$2 AND b %[1]v \$3\).*ORDER BY a %[2]v, b %[2]v`,
				testCase.expectedOperator,
				testCase.expectedOrder,
			),
		).WillReturnRows(getDummyRows())

		var dummy Dummy

		p := newDummyPaginatorWithAfterCursor()

		p.SetOrder(testCase.order)

		p.Paginate(db, &dummy)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expections: %s", err)
	}
}

func newDummyPaginatorWithBeforeCursor() Paginator {
	p := New()
	p.SetKeys("A", "B")
	p.SetBeforeCursor(p.encode(reflect.ValueOf(Dummy{"A", "B"})))
	return p
}

func TestPaginateWithBeforeCursor(t *testing.T) {
	mock, db := mockDB()

	for _, testCase := range []struct {
		order            order
		expectedOperator string
		expectedOrder    order
	}{
		{ASC, "<", DESC},
		{DESC, ">", ASC},
	} {
		mock.ExpectQuery(
			fmt.Sprintf(
				`.*WHERE.*\(a %[1]v \$1 OR a = \$2 AND b %[1]v \$3\).*ORDER BY a %[2]v, b %[2]v`,
				testCase.expectedOperator,
				testCase.expectedOrder,
			),
		).WillReturnRows(getDummyRows())

		var dummy Dummy

		p := newDummyPaginatorWithBeforeCursor()

		p.SetOrder(testCase.order)

		p.Paginate(db, &dummy)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expections: %s", err)
	}
}

type IDModel struct {
	ID string
}

func TestPostProcessOutcomeManipulation(t *testing.T) {
	for index, testCase := range []struct {
		afterCursor  string
		beforeCursor string
		limit        int
		out          []IDModel
		expected     []IDModel
	}{
		// after cursor is provided and is under limit
		{"after", "", 5, []IDModel{{"A"}, {"B"}, {"C"}}, []IDModel{{"A"}, {"B"}, {"C"}}},

		// before cursor is provided and is under limit
		{"", "before", 5, []IDModel{{"A"}, {"B"}, {"C"}}, []IDModel{{"C"}, {"B"}, {"A"}}},

		// both cursors are not provided and is under limit
		{"", "", 5, []IDModel{{"A"}, {"B"}, {"C"}}, []IDModel{{"A"}, {"B"}, {"C"}}},

		// after cursor is provided and has more data
		{"after", "", 2, []IDModel{{"A"}, {"B"}, {"C"}}, []IDModel{{"A"}, {"B"}}},

		// before cursor is provided and has more data
		{"", "before", 2, []IDModel{{"A"}, {"B"}, {"C"}}, []IDModel{{"B"}, {"A"}}},

		// none of cursors are provided and has more data
		{"", "", 2, []IDModel{{"A"}, {"B"}, {"C"}}, []IDModel{{"A"}, {"B"}}},
	} {
		p := New()

		p.SetKeys("ID")
		p.SetLimit(testCase.limit)
		p.SetAfterCursor(testCase.afterCursor)
		p.SetBeforeCursor(testCase.beforeCursor)

		p.postProcess(&testCase.out)

		if !reflect.DeepEqual(testCase.out, testCase.expected) {
			t.Errorf("[case %v] expected `%v`, but got `%v`", index+1, testCase.expected, testCase.out)
		}
	}
}

func parseCursor(p Paginator, cursor string) string {
	if cursor == "" {
		return ""
	}
	return combine(p.decode(cursor))
}

func combine(cursors []interface{}) string {
	strs := make([]string, len(cursors))

	for index, cursor := range cursors {
		strs[index] = fmt.Sprintf("%v", cursor)
	}
	return strings.Join(strs, ",")
}

func TestPostProcessCursorsGenreation(t *testing.T) {
	for index, testCase := range []struct {
		afterCursor              string
		beforeCursor             string
		limit                    int
		out                      []IDModel
		expectedNextAfterCursor  string
		expectedNextBeforeCursor string
	}{
		// after cursor is provided and is under limit
		{"after", "", 5, []IDModel{{"A"}, {"B"}, {"C"}}, "", "A"},

		// before cursor is provided and is under limit
		{"", "before", 5, []IDModel{{"A"}, {"B"}, {"C"}}, "A", ""},

		// none of cursors are provided and is under limit
		{"", "", 5, []IDModel{{"A"}, {"B"}, {"C"}}, "", ""},

		// after cursor is provided and has more data
		{"after", "", 2, []IDModel{{"A"}, {"B"}, {"C"}}, "B", "A"},

		// before cursor is provided and has more data
		{"", "before", 2, []IDModel{{"A"}, {"B"}, {"C"}}, "A", "B"},

		// none of cursors are provided and has more data
		{"", "", 2, []IDModel{{"A"}, {"B"}, {"C"}}, "B", ""},
	} {
		p := New()

		p.SetKeys("ID")
		p.SetLimit(testCase.limit)
		p.SetAfterCursor(testCase.afterCursor)
		p.SetBeforeCursor(testCase.beforeCursor)

		p.postProcess(&testCase.out)

		nextCursors := p.GetNextCursors()
		nextAfterCursor := parseCursor(p, nextCursors.AfterCursor)
		nextBeforeCursor := parseCursor(p, nextCursors.BeforeCursor)

		assertEqual(t, fmt.Sprintf("%v: next after cursor", index+1), nextAfterCursor, testCase.expectedNextAfterCursor)
		assertEqual(t, fmt.Sprintf("%v: next before cursor", index+1), nextBeforeCursor, testCase.expectedNextBeforeCursor)
	}
}

type CompositeModel struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func TestPostProcessModelEncoding(t *testing.T) {
	for index, testCase := range []struct {
		keys []string
	}{
		{[]string{"ID"}},
		{[]string{"CreatedAt", "ID"}},
		{[]string{"UpdatedAt", "CreatedAt", "ID"}},
	} {
		compositeModels := []CompositeModel{
			{"ID", time.Now().UTC(), time.Now().UTC().Add(time.Hour * time.Duration(1))},
			{"__", time.Now().UTC(), time.Now().UTC()},
		}
		p := New()

		p.SetKeys(testCase.keys...)
		p.SetLimit(0)

		p.postProcess(&compositeModels)

		fields := p.decode(
			p.GetNextCursors().AfterCursor,
		)
		for i, key := range testCase.keys {
			got := fields[i]
			expected := reflect.ValueOf(compositeModels[0]).FieldByName(key).Interface()

			assertEqual(t, index+1, got, expected)
		}
	}
}

func TestGetCursorQueryTemplate(t *testing.T) {
	for index, testCase := range []struct {
		keys     []string
		operator string
		expected string
	}{
		// single key
		{[]string{"ID"}, ">", "id > ?"},
		{[]string{"CreatedAt"}, "<", "created_at < ?"},

		// multiple keys
		{[]string{"CreatedAt", "ID"}, ">", "created_at > ? OR created_at = ? AND id > ?"},
		{[]string{"CreatedAt", "UpdatedAt", "ID"}, "<", "created_at < ? OR created_at = ? AND updated_at < ? OR created_at = ? AND updated_at = ? AND id < ?"},
	} {
		p := New()

		p.SetKeys(testCase.keys...)

		got := p.getCursorQueryTemplate(testCase.operator)

		assertEqual(t, index+1, got, testCase.expected)
	}
}

func TestGetCursorQueryArgs(t *testing.T) {
	for index, testCase := range []struct {
		cursors  []interface{}
		expected string
	}{
		{[]interface{}{"a"}, "a"},
		{[]interface{}{"a", "b"}, "a,a,b"},
		{[]interface{}{"a", "b", "c"}, "a,a,b,a,b,c"},
		{[]interface{}{"a", "b", "c", 1, 2, 3}, "a,a,b,a,b,c,a,b,c,1,a,b,c,1,2,a,b,c,1,2,3"},
	} {
		p := New()

		got := combine(
			p.getCursorQueryArgs(testCase.cursors),
		)
		assertEqual(t, index+1, got, testCase.expected)
	}
}

func TestGetOperator(t *testing.T) {
	for index, testCase := range []struct {
		afterCursor  string
		beforeCursor string
		order        order
		expected     string
	}{
		// after cursor with different order
		{"after", "", ASC, ">"},
		{"after", "", DESC, "<"},

		// before cursor with different order
		{"", "before", ASC, "<"},
		{"", "before", DESC, ">"},

		// both cursors are provided
		{"after", "before", ASC, ">"},
		{"after", "before", DESC, "<"},

		// none of cursors are provided
		{"", "", ASC, "="},
		{"", "", DESC, "="},
	} {
		p := New()

		p.SetAfterCursor(testCase.afterCursor)
		p.SetBeforeCursor(testCase.beforeCursor)
		p.SetOrder(testCase.order)

		got := p.getOperator()

		assertEqual(t, index+1, got, testCase.expected)
	}
}

func TestGetOrderWithMultipleKeys(t *testing.T) {
	for index, testCase := range []struct {
		keys     []string
		order    order
		expected string
	}{
		// single key
		{[]string{"ID"}, ASC, "id ASC"},
		{[]string{"CreatedAt"}, DESC, "created_at DESC"},

		// multiple keys
		{[]string{"CreatedAt", "ID"}, ASC, "created_at ASC, id ASC"},
		{[]string{"CreatedAt", "UpdatedAt", "ID"}, DESC, "created_at DESC, updated_at DESC, id DESC"},
	} {
		p := New()

		p.SetKeys(testCase.keys...)
		p.SetOrder(testCase.order)

		got := p.getOrder()

		assertEqual(t, index+1, got, testCase.expected)
	}
}

func TestGetOrderWithCursors(t *testing.T) {
	for index, testCase := range []struct {
		afterCursor  string
		beforeCursor string
		order        order
		expected     string
	}{
		// after cursor is provided
		{"after", "", ASC, "id ASC"},
		{"after", "", DESC, "id DESC"},

		// before cursor is provided
		{"", "before", ASC, "id DESC"},
		{"", "before", DESC, "id ASC"},

		// both cursors are provided
		{"after", "before", ASC, "id ASC"},
		{"after", "before", DESC, "id DESC"},

		// none of cursors are provided
		{"", "", ASC, "id ASC"},
		{"", "", DESC, "id DESC"},
	} {
		p := New()

		p.SetKeys("ID")
		p.SetOrder(testCase.order)
		p.SetAfterCursor(testCase.afterCursor)
		p.SetBeforeCursor(testCase.beforeCursor)

		got := p.getOrder()

		assertEqual(t, index+1, got, testCase.expected)
	}
}
