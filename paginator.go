package paginator

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
)

type order string

const (
	// ASC refers to ascending order
	ASC order = "ASC"
	// DESC refers to descending order
	DESC order = "DESC"
)

const (
	defaultPageCount = 10
	defaultOrder     = DESC
)

// New inits paginator
func New() Paginator {
	return Paginator{}
}

// Paginator a builder doing pagination
type Paginator struct {
	afterCursor      string
	beforeCursor     string
	nextAfterCursor  string
	nextBeforeCursor string
	keys             []string
	limit            int
	order            order
}

// Cursors groups after/before cursors
type Cursors struct {
	AfterCursor  string `json:"afterCursor"`
	BeforeCursor string `json:"beforeCursor"`
}

// SetAfterCursor sets paging after cursor
func (p *Paginator) SetAfterCursor(cursor string) {
	p.afterCursor = cursor
}

// SetBeforeCursor sets paging before cursor
func (p *Paginator) SetBeforeCursor(cursor string) {
	p.beforeCursor = cursor
}

// SetKeys sets paging keys
func (p *Paginator) SetKeys(keys ...string) {
	p.keys = append(p.keys, keys...)
}

// SetLimit sets paging limit
func (p *Paginator) SetLimit(limit int) {
	p.limit = limit
}

// SetOrder sets paging order
func (p *Paginator) SetOrder(order order) {
	p.order = order
}

// GetNextCursors gets new cursors after pagination
func (p *Paginator) GetNextCursors() Cursors {
	return Cursors{
		AfterCursor:  p.nextAfterCursor,
		BeforeCursor: p.nextBeforeCursor,
	}
}

// Paginate paginates data
func (p *Paginator) Paginate(stmt *gorm.DB, out interface{}) *gorm.DB {
	p.initOptions()

	result := p.appendPagingQuery(stmt).Find(out)
	// out must be a pointer or gorm will panic above
	if reflect.ValueOf(out).Elem().Type().Kind() == reflect.Slice && reflect.ValueOf(out).Elem().Len() > 0 {
		p.postProcess(out)
	}
	return result
}

/* private */

func (p *Paginator) initOptions() {
	if len(p.keys) == 0 {
		p.keys = append(p.keys, "ID")
	}
	if p.limit == 0 {
		p.limit = defaultPageCount
	}
	if p.order == "" {
		p.order = defaultOrder
	}
}

func (p *Paginator) appendPagingQuery(stmt *gorm.DB) *gorm.DB {
	var cursors []interface{}

	if p.hasAfterCursor() {
		cursors = p.decode(p.afterCursor)
	} else if p.hasBeforeCursor() {
		cursors = p.decode(p.beforeCursor)
	}
	if len(cursors) > 0 {
		stmt = stmt.Where(
			p.getCursorQueryTemplate(p.getOperator()),
			p.getCursorQueryArgs(cursors)...,
		)
	}
	stmt = stmt.Limit(p.limit + 1)
	stmt = stmt.Order(p.getOrder())

	return stmt
}

func (p *Paginator) hasAfterCursor() bool {
	return p.afterCursor != ""
}

func (p *Paginator) hasBeforeCursor() bool {
	return p.beforeCursor != ""
}

func (p *Paginator) decode(cursor string) []interface{} {
	bytes, err := base64.StdEncoding.DecodeString(cursor)

	if err != nil {
		return nil
	}
	fields := strings.Split(string(bytes[:]), ",")
	keys := make([]interface{}, len(fields))

	for index, field := range fields {
		keys[index] = deconvert(field)
	}
	return keys
}

func (p *Paginator) getCursorQueryTemplate(operator string) string {
	queries := make([]string, len(p.keys))
	composite := ""

	for index, key := range p.keys {
		snakeKey := strcase.ToSnake(key)
		queries[index] = fmt.Sprintf("%s%s %s ?", composite, snakeKey, operator)
		composite = fmt.Sprintf("%s%s = ? AND ", composite, snakeKey)
	}
	return strings.Join(queries, " OR ")
}

func (p *Paginator) getCursorQueryArgs(cursors []interface{}) (args []interface{}) {
	for i := 1; i <= len(cursors); i++ {
		args = append(args, cursors[:i]...)
	}
	return
}

func (p *Paginator) getOperator() string {
	if p.hasAfterCursor() {
		if p.order == ASC {
			return ">"
		}
		return "<"
	}
	if p.hasBeforeCursor() {
		if p.order == ASC {
			return "<"
		}
		return ">"
	}
	return "="
}

func (p *Paginator) getOrder() string {
	order := p.order

	if !p.hasAfterCursor() && p.hasBeforeCursor() {
		order = flip(p.order)
	}
	orders := make([]string, len(p.keys))

	for index, key := range p.keys {
		orders[index] = fmt.Sprintf("%s %s", strcase.ToSnake(key), order)
	}
	return strings.Join(orders, ", ")
}

func (p *Paginator) postProcess(out interface{}) {
	elems := reflect.ValueOf(out).Elem()
	// has more data after/before given cursor
	hasMore := elems.Len() > p.limit

	if hasMore {
		elems.Set(elems.Slice(0, elems.Len()-1))
	}
	// reverse out in before cursor scenario
	if !p.hasAfterCursor() && p.hasBeforeCursor() {
		elems.Set(reverse(elems))
	}
	if p.hasBeforeCursor() || hasMore {
		p.nextAfterCursor = p.encode(elems.Index(elems.Len() - 1))
	}
	if p.hasAfterCursor() || (hasMore && p.hasBeforeCursor()) {
		p.nextBeforeCursor = p.encode(elems.Index(0))
	}
	return
}

func (p *Paginator) encode(v reflect.Value) string {
	fields := make([]string, len(p.keys))

	for index, key := range p.keys {
		fields[index] = convert(v.FieldByName(key).Interface())
	}
	return base64.StdEncoding.EncodeToString(
		[]byte(strings.Join(fields, ",")),
	)
}
