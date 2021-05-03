package paginator

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pilagod/gorm-cursor-paginator/cursor"
	"gorm.io/gorm"
)

// New creates paginator
func New(opts ...Option) *Paginator {
	p := &Paginator{}
	for _, opt := range append([]Option{&defaultConfig}, opts...) {
		opt.Apply(p)
	}
	return p
}

// Paginator a builder doing pagination
type Paginator struct {
	cursor    cursor.Cursor
	keys      []string
	tableKeys []string
	limit     int
	order     Order
}

// SetKeys sets paging keys
func (p *Paginator) SetKeys(keys ...string) {
	p.keys = keys
}

// SetLimit sets paging limit
func (p *Paginator) SetLimit(limit int) {
	p.limit = limit
}

// SetOrder sets paging order
func (p *Paginator) SetOrder(order Order) {
	p.order = order
}

// SetAfterCursor sets paging after cursor
func (p *Paginator) SetAfterCursor(afterCursor string) {
	p.cursor.After = &afterCursor
}

// SetBeforeCursor sets paging before cursor
func (p *Paginator) SetBeforeCursor(beforeCursor string) {
	p.cursor.Before = &beforeCursor
}

// Paginate paginates data
func (p *Paginator) Paginate(db *gorm.DB, out interface{}) (result *gorm.DB, c cursor.Cursor, err error) {
	result = db.WithContext(context.Background())
	p.init(result, out)
	// decode cursor
	fields, err := p.decodeCursor(out)
	if err != nil {
		return
	}
	result = p.appendPagingQuery(result, fields).Find(out)
	// out must be a pointer or gorm will panic above
	elems := reflect.ValueOf(out).Elem()
	// only encode next cursor when elems is not empty slice
	if elems.Kind() == reflect.Slice && elems.Len() > 0 {
		hasMore := elems.Len() > p.limit
		if hasMore {
			elems.Set(elems.Slice(0, elems.Len()-1))
		}
		if p.isBackward() {
			elems.Set(reverse(elems))
		}
		if c, err = p.encodeCursor(elems, hasMore); err != nil {
			return
		}
	}
	return
}

/* private */

func (p *Paginator) init(db *gorm.DB, out interface{}) {
	// https://stackoverflow.com/questions/51999441/how-to-get-a-table-name-from-a-model-in-gorm
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(out)
	table := stmt.Schema.Table
	for _, key := range p.keys {
		p.tableKeys = append(p.tableKeys, fmt.Sprintf("%s.%s", table, strcase.ToSnake(key)))
	}
}

func (p *Paginator) decodeCursor(out interface{}) ([]interface{}, error) {
	decoder, err := cursor.NewDecoder(out, p.keys...)
	if err != nil {
		return nil, err
	}
	if p.isForward() {
		return decoder.Decode(*p.cursor.After)
	}
	if p.isBackward() {
		return decoder.Decode(*p.cursor.Before)
	}
	return nil, nil
}

func (p *Paginator) isForward() bool {
	return p.cursor.After != nil
}

func (p *Paginator) isBackward() bool {
	// forward take precedence over backward
	return !p.isForward() && p.cursor.Before != nil
}

func (p *Paginator) appendPagingQuery(db *gorm.DB, fields []interface{}) *gorm.DB {
	stmt := db
	stmt = stmt.Limit(p.limit + 1)
	stmt = stmt.Order(p.getOrder())
	if len(fields) > 0 {
		stmt = stmt.Where(
			p.getCursorQuery(),
			p.getCursorQueryArgs(fields)...,
		)
	}
	return stmt
}

func (p *Paginator) getOrder() string {
	order := p.order
	if p.isBackward() {
		order = p.order.Flip()
	}
	orders := make([]string, len(p.tableKeys))
	for index, sqlKey := range p.tableKeys {
		orders[index] = fmt.Sprintf("%s %s", sqlKey, order)
	}
	return strings.Join(orders, ", ")
}

func (p *Paginator) getCursorQuery() string {
	qs := make([]string, len(p.tableKeys))
	op := p.getOperator()
	composite := ""
	for i, sqlKey := range p.tableKeys {
		qs[i] = fmt.Sprintf("%s%s %s ?", composite, sqlKey, op)
		composite = fmt.Sprintf("%s%s = ? AND ", composite, sqlKey)
	}
	return strings.Join(qs, " OR ")
}

func (p *Paginator) getOperator() string {
	if (p.isForward() && p.order == ASC) ||
		(p.isBackward() && p.order == DESC) {
		return ">"
	}
	return "<"
}

func (p *Paginator) getCursorQueryArgs(fields []interface{}) (args []interface{}) {
	for i := 1; i <= len(fields); i++ {
		args = append(args, fields[:i]...)
	}
	return
}

func (p *Paginator) encodeCursor(elems reflect.Value, hasMore bool) (result cursor.Cursor, err error) {
	encoder := cursor.NewEncoder(p.keys...)
	// encode after cursor
	if p.isBackward() || hasMore {
		c, err := encoder.Encode(elems.Index(elems.Len() - 1))
		if err != nil {
			return cursor.Cursor{}, err
		}
		result.After = &c
	}
	// encode before cursor
	if p.isForward() || (hasMore && p.isBackward()) {
		c, err := encoder.Encode(elems.Index(0))
		if err != nil {
			return cursor.Cursor{}, err
		}
		result.Before = &c
	}
	return
}
