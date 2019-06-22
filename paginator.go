package paginator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
)

// Order type for order
type Order string

// Orders
const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

const (
	defaultLimit = 10
	defaultOrder = DESC
)

// New inits paginator
func New() *Paginator {
	return &Paginator{}
}

// Paginator a builder doing pagination
type Paginator struct {
	cursor  Cursor
	next    Cursor
	keys    []string
	sqlKeys []string
	limit   int
	order   Order
}

// Cursor cursor data
type Cursor struct {
	After  *string `json:"after"`
	Before *string `json:"before"`
}

// SetAfterCursor sets paging after cursor
func (p *Paginator) SetAfterCursor(afterCursor string) {
	p.cursor.After = &afterCursor
}

// SetBeforeCursor sets paging before cursor
func (p *Paginator) SetBeforeCursor(beforeCursor string) {
	p.cursor.Before = &beforeCursor
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
func (p *Paginator) SetOrder(order Order) {
	p.order = order
}

// GetNextCursor returns cursor for next pagination
func (p *Paginator) GetNextCursor() Cursor {
	return p.next
}

// Paginate paginates data
func (p *Paginator) Paginate(stmt *gorm.DB, out interface{}) *gorm.DB {
	p.initOptions()
	p.initModelInfo(stmt, out)
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
		p.limit = defaultLimit
	}
	if p.order == "" {
		p.order = defaultOrder
	}
}

func (p *Paginator) initModelInfo(db *gorm.DB, out interface{}) {
	table := db.NewScope(out).TableName()
	for _, key := range p.keys {
		p.sqlKeys = append(p.sqlKeys, fmt.Sprintf("%s.%s", table, strcase.ToSnake(key)))
	}
}

func (p *Paginator) appendPagingQuery(stmt *gorm.DB) *gorm.DB {
	var fields []interface{}
	if p.hasAfterCursor() {
		fields = p.decode(*p.cursor.After)
	} else if p.hasBeforeCursor() {
		fields = p.decode(*p.cursor.Before)
	}
	if len(fields) > 0 {
		stmt = stmt.Where(
			p.getCursorQuery(p.getOperator()),
			p.getCursorQueryArgs(fields)...,
		)
	}
	stmt = stmt.Limit(p.limit + 1)
	stmt = stmt.Order(p.getOrder())
	return stmt
}

func (p *Paginator) getCursorQuery(operator string) string {
	queries := make([]string, len(p.sqlKeys))
	composite := ""
	for index, sqlKey := range p.sqlKeys {
		queries[index] = fmt.Sprintf("%s%s %s ?", composite, sqlKey, operator)
		composite = fmt.Sprintf("%s%s = ? AND ", composite, sqlKey)
	}
	return strings.Join(queries, " OR ")
}

func (p *Paginator) getCursorQueryArgs(fields []interface{}) (args []interface{}) {
	for i := 1; i <= len(fields); i++ {
		args = append(args, fields[:i]...)
	}
	return
}

func (p *Paginator) getOperator() string {
	if (p.hasAfterCursor() && p.order == ASC) ||
		(p.hasBeforeCursorOnly() && p.order == DESC) {
		return ">"
	}
	return "<"
}

func (p *Paginator) getOrder() string {
	order := p.order
	if p.hasBeforeCursorOnly() {
		order = flip(p.order)
	}
	orders := make([]string, len(p.sqlKeys))
	for index, sqlKey := range p.sqlKeys {
		orders[index] = fmt.Sprintf("%s %s", sqlKey, order)
	}
	return strings.Join(orders, ", ")
}

func (p *Paginator) postProcess(out interface{}) {
	elems := reflect.ValueOf(out).Elem()
	hasMore := elems.Len() > p.limit
	if hasMore {
		elems.Set(elems.Slice(0, elems.Len()-1))
	}
	if p.hasBeforeCursorOnly() {
		elems.Set(reverse(elems))
	}
	if p.hasBeforeCursorOnly() || hasMore {
		cursor := p.encode(elems.Index(elems.Len() - 1))
		p.next.After = &cursor
	}
	if p.hasAfterCursor() || (hasMore && p.hasBeforeCursorOnly()) {
		cursor := p.encode(elems.Index(0))
		p.next.Before = &cursor
	}
	return
}

func (p *Paginator) encode(v reflect.Value) string {
	fields := make([]string, len(p.keys))
	for index, key := range p.keys {
		fields[index] = convert(v.FieldByName(key).Interface())
	}
	return encodeBase64(fields)
}

func (p *Paginator) decode(cursor string) []interface{} {
	fieldsWithType := decodeBase64(cursor)
	fields := make([]interface{}, len(fieldsWithType))
	for index, fieldWithType := range fieldsWithType {
		fields[index] = revert(fieldWithType)
	}
	return fields
}

func (p *Paginator) hasAfterCursor() bool {
	return p.cursor.After != nil
}

func (p *Paginator) hasBeforeCursor() bool {
	return p.cursor.Before != nil
}

func (p *Paginator) hasBeforeCursorOnly() bool {
	return !p.hasAfterCursor() && p.hasBeforeCursor()
}
