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
	cursor cursor.Cursor
	rules  []Rule
	limit  int
	order  Order
}

func (p *Paginator) SetRules(rules ...Rule) {
	p.rules = make([]Rule, len(rules))
	copy(p.rules, rules)
}

// SetKeys sets paging keys
func (p *Paginator) SetKeys(keys ...string) {
	rules := make([]Rule, len(keys))
	for i, key := range keys {
		rules[i] = Rule{
			Key: key,
		}
	}
	p.SetRules(rules...)
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
	if err = p.validate(); err != nil {
		return
	}
	dbCtx := db.WithContext(context.Background())
	p.setup(dbCtx, out)
	// decode cursor
	fields, err := p.decodeCursor(out)
	if err != nil {
		return
	}
	result = p.appendPagingQuery(dbCtx, fields).Find(out)
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

func (p *Paginator) validate() (err error) {
	if err = p.order.Validate(false); err != nil {
		return
	}
	for _, rule := range p.rules {
		if err = rule.Validate(); err != nil {
			return
		}
	}
	return
}

func (p *Paginator) setup(db *gorm.DB, out interface{}) {
	var outTable string
	for i := range p.rules {
		if p.rules[i].SQLRepr == "" {
			if outTable == "" {
				// https://stackoverflow.com/questions/51999441/how-to-get-a-table-name-from-a-model-in-gorm
				stmt := &gorm.Statement{DB: db}
				stmt.Parse(out)
				outTable = stmt.Schema.Table
			}
			// TODO: get gorm column tag
			p.rules[i].SQLRepr = fmt.Sprintf("%s.%s", outTable, strcase.ToSnake(p.rules[i].Key))
		}
		if p.rules[i].Order == "" {
			p.rules[i].Order = p.order
		}
	}
}

func (p *Paginator) decodeCursor(out interface{}) ([]interface{}, error) {
	decoder, err := cursor.NewDecoder(out, p.getKeys()...)
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
	stmt = stmt.Order(p.buildOrderSQL())
	if len(fields) > 0 {
		stmt = stmt.Where(
			p.buildCursorSQLQuery(),
			p.buildCursorSQLQueryArgs(fields)...,
		)
	}
	return stmt
}

func (p *Paginator) buildOrderSQL() string {
	orders := make([]string, len(p.rules))
	for i, rule := range p.rules {
		order := rule.Order
		if p.isBackward() {
			order = p.order.Flip()
		}
		orders[i] = fmt.Sprintf("%s %s", rule.SQLRepr, order)
	}
	return strings.Join(orders, ", ")
}

func (p *Paginator) buildCursorSQLQuery() string {
	queries := make([]string, len(p.rules))
	query := ""
	for i, rule := range p.rules {
		operator := "<"
		if (p.isForward() && rule.Order == ASC) ||
			(p.isBackward() && rule.Order == DESC) {
			operator = ">"
		}
		queries[i] = fmt.Sprintf("%s%s %s ?", query, rule.SQLRepr, operator)
		query = fmt.Sprintf("%s%s = ? AND ", query, rule.SQLRepr)
	}
	// for exmaple:
	// a > 1 OR a = 1 AND b > 2 OR a = 1 AND b = 2 AND c > 3
	return strings.Join(queries, " OR ")
}

func (p *Paginator) buildCursorSQLQueryArgs(fields []interface{}) (args []interface{}) {
	for i := 1; i <= len(fields); i++ {
		args = append(args, fields[:i]...)
	}
	return
}

func (p *Paginator) encodeCursor(elems reflect.Value, hasMore bool) (result cursor.Cursor, err error) {
	encoder := cursor.NewEncoder(p.getKeys()...)
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

/* rules */

func (p *Paginator) getKeys() []string {
	keys := make([]string, len(p.rules))
	for i, rule := range p.rules {
		keys[i] = rule.Key
	}
	return keys
}
