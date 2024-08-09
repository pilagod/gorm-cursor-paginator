package paginator

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"

	"github.com/pilagod/gorm-cursor-paginator/v2/cursor"
	"github.com/pilagod/gorm-cursor-paginator/v2/internal/util"
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
	cursor        Cursor
	rules         []Rule
	limit         int
	order         Order
	allowTupleCmp bool
	cursorCodec   CursorCodec
}

// SetRules sets paging rules
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

// SetAllowTupleCmp enables or disables tuple comparison optimization
func (p *Paginator) SetAllowTupleCmp(allow bool) {
	p.allowTupleCmp = allow
}

// SetCursorCodec sets custom cursor codec
func (p *Paginator) SetCursorCodec(codec CursorCodec) {
	p.cursorCodec = codec
}

// Paginate paginates data
func (p *Paginator) Paginate(db *gorm.DB, dest interface{}) (result *gorm.DB, c Cursor, err error) {
	if err = p.validate(db, dest); err != nil {
		return
	}
	if err = p.setup(db, dest); err != nil {
		return
	}
	fields, err := p.decodeCursor(dest)
	if err != nil {
		return
	}
	if result = p.appendPagingQuery(db, fields).Find(dest); result.Error != nil {
		return
	}
	// dest must be a pointer type or gorm will panic above
	elems := reflect.ValueOf(dest).Elem()
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

func (p *Paginator) validate(db *gorm.DB, dest interface{}) (err error) {
	if len(p.rules) == 0 {
		return ErrNoRule
	}
	if p.limit <= 0 {
		return ErrInvalidLimit
	}
	if err = p.order.validate(); err != nil {
		return
	}
	for _, rule := range p.rules {
		if err = rule.validate(db, dest); err != nil {
			return
		}
	}
	return
}

func (p *Paginator) setup(db *gorm.DB, dest interface{}) error {
	var sqlTable string
	for i := range p.rules {
		rule := &p.rules[i]
		if rule.SQLRepr == "" {
			if sqlTable == "" {
				schema, err := util.ParseSchema(db, dest)
				if err != nil {
					return err
				}
				sqlTable = schema.Table
			}
			sqlKey := p.parseSQLKey(db, dest, rule.Key)
			rule.SQLRepr = fmt.Sprintf("%s.%s", sqlTable, sqlKey)
		}

		if rule.NULLReplacement != nil {
			nullReplacement := fmt.Sprintf("'%v'", rule.NULLReplacement)
			if rule.SQLType != nil {
				nullReplacement = fmt.Sprintf("CAST(%s as %s)", nullReplacement, *rule.SQLType)
			}
			rule.SQLRepr = fmt.Sprintf("COALESCE(%s, %s)", rule.SQLRepr, nullReplacement)
		}
		// cast to the underlying SQL type
		if rule.SQLType != nil {
			rule.SQLRepr = fmt.Sprintf("CAST( %s AS %s )", rule.SQLRepr, *rule.SQLType)
		}
		if rule.Order == "" {
			rule.Order = p.order
		}
	}
	return nil
}

func (p *Paginator) parseSQLKey(db *gorm.DB, dest interface{}, key string) string {
	// dest is already validated at validataion phase
	schema, _ := util.ParseSchema(db, dest)
	return schema.LookUpField(key).DBName
}

// https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	// reflect.Array is intentionally omitted since calling IsNil() on the value
	// of an array will panic
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func (p *Paginator) decodeCursor(dest interface{}) (result []interface{}, err error) {
	if p.isForward() {
		if result, err = p.cursorCodec.Decode(p.getDecoderFields(), *p.cursor.After, dest); err != nil {
			err = ErrInvalidCursor
		}
	} else if p.isBackward() {
		if result, err = p.cursorCodec.Decode(p.getDecoderFields(), *p.cursor.Before, dest); err != nil {
			err = ErrInvalidCursor
		}
	}
	// replace null values
	for i := range result {
		value := result[i]
		// for custom types, evaluate isNil on the underlying value
		if ct, ok := result[i].(cursor.CustomType); ok && p.rules[i].CustomType != nil {
			value, err = ct.GetCustomTypeValue(p.rules[i].CustomType.Meta)
		}
		if isNil(value) {
			result[i] = p.rules[i].NULLReplacement
		}
	}
	return
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

	if len(fields) == 0 {
		return stmt
	}

	if p.allowTupleCmp && p.canOptimizePagingQuery() {
		return stmt.Where(p.buildOptimizedCursorSQLQuery(), fields)
	}

	return stmt.Where(
		p.buildCursorSQLQuery(),
		p.buildCursorSQLQueryArgs(fields)...,
	)
}

func (p *Paginator) buildOrderSQL() string {
	orders := make([]string, len(p.rules))
	for i, rule := range p.rules {
		order := rule.Order
		if p.isBackward() {
			order = order.flip()
		}
		orders[i] = fmt.Sprintf("%s %s", rule.SQLRepr, order)
	}
	return strings.Join(orders, ", ")
}

func (p *Paginator) buildCursorSQLQuery() string {
	queries := make([]string, len(p.rules))
	query := ""
	for i, rule := range p.rules {
		operator := p.getCmpOperator(rule.Order)
		queries[i] = fmt.Sprintf("%s%s %s ?", query, rule.SQLRepr, operator)
		query = fmt.Sprintf("%s%s = ? AND ", query, rule.SQLRepr)
	}
	// for exmaple:
	// a > 1 OR a = 1 AND b > 2 OR a = 1 AND b = 2 AND c > 3
	return strings.Join(queries, " OR ")
}

// We can only optimize paging query if sorting orders are consistent across
// all columns used in cursor. This is a prerequisite for tuple comparison that
// optimized queries use.
func (p *Paginator) canOptimizePagingQuery() bool {
	order := p.rules[0].Order

	for _, rule := range p.rules {
		if order != rule.Order {
			return false
		}
	}

	return true
}

func (p *Paginator) getCmpOperator(order Order) string {
	if (p.isForward() && order == ASC) || (p.isBackward() && order == DESC) {
		return ">"
	}

	return "<"
}

func (p *Paginator) buildOptimizedCursorSQLQuery() string {
	names := make([]string, len(p.rules))

	for i, rule := range p.rules {
		names[i] = rule.SQLRepr
	}

	return fmt.Sprintf(
		"(%s) %s ?",
		strings.Join(names, ", "),
		p.getCmpOperator(p.rules[0].Order),
	)
}

func (p *Paginator) buildCursorSQLQueryArgs(fields []interface{}) (args []interface{}) {
	for i := 1; i <= len(fields); i++ {
		args = append(args, fields[:i]...)
	}
	return
}

func (p *Paginator) encodeCursor(elems reflect.Value, hasMore bool) (result Cursor, err error) {
	// encode after cursor
	if p.isBackward() || hasMore {
		c, err := p.cursorCodec.Encode(p.getEncoderFields(), elems.Index(elems.Len()-1))
		if err != nil {
			return Cursor{}, err
		}
		result.After = &c
	}
	// encode before cursor
	if p.isForward() || (hasMore && p.isBackward()) {
		c, err := p.cursorCodec.Encode(p.getEncoderFields(), elems.Index(0))
		if err != nil {
			return Cursor{}, err
		}
		result.Before = &c
	}
	return
}

/* custom types */

func (p *Paginator) getEncoderFields() []cursor.EncoderField {
	fields := make([]cursor.EncoderField, len(p.rules))
	for i, rule := range p.rules {
		fields[i].Key = rule.Key
		if rule.CustomType != nil {
			fields[i].Meta = rule.CustomType.Meta
		}
	}
	return fields
}

func (p *Paginator) getDecoderFields() []cursor.DecoderField {
	fields := make([]cursor.DecoderField, len(p.rules))
	for i, rule := range p.rules {
		fields[i].Key = rule.Key
		if rule.CustomType != nil {
			fields[i].Type = &rule.CustomType.Type
		}
	}
	return fields
}
