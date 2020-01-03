package paginator

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
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
	cursor   Cursor
	next     Cursor
	keys     []string
	keyKinds []kind
	sqlKeys  []string
	limit    int
	order    Order
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
	p.initKeyKinds(out)
	p.initTableKeys(stmt, out)
	result := p.appendPagingQuery(stmt).Find(out)
	// out must be a pointer or gorm will panic above
	elems := reflect.ValueOf(out).Elem()
	if elems.Kind() == reflect.Slice && elems.Len() > 0 {
		p.postProcess(out)
	}
	return result
}

// Encode encodes fields to cursor
func (p *Paginator) Encode(v interface{}) string {
	return base64.StdEncoding.EncodeToString(p.marshalJSON(v))
}

// Decode decodes cursor to fields
func (p *Paginator) Decode(cursor string) []interface{} {
	b, err := base64.StdEncoding.DecodeString(cursor)
	// @TODO: return proper error
	if err != nil {
		return nil
	}
	return p.unmarshalJSON(b)
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

func (p *Paginator) initKeyKinds(out interface{}) {
	rt := reflect.ValueOf(out).Type()
	for rt.Kind() == reflect.Slice || rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		// element of out must be struct, if not, just pass it to gorm to handle the error
		return
	}
	p.keyKinds = make([]kind, len(p.keys))
	for i, key := range p.keys {
		field, _ := rt.FieldByName(key)
		p.keyKinds[i] = toKind(field.Type)
	}
}

func (p *Paginator) initTableKeys(db *gorm.DB, out interface{}) {
	table := db.NewScope(out).TableName()
	for _, key := range p.keys {
		p.sqlKeys = append(p.sqlKeys, fmt.Sprintf("%s.%s", table, strcase.ToSnake(key)))
	}
}

func (p *Paginator) appendPagingQuery(stmt *gorm.DB) *gorm.DB {
	var fields []interface{}
	if p.hasAfterCursor() {
		fields = p.Decode(*p.cursor.After)
	} else if p.hasBeforeCursor() {
		fields = p.Decode(*p.cursor.Before)
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

func (p *Paginator) hasAfterCursor() bool {
	return p.cursor.After != nil
}

func (p *Paginator) hasBeforeCursor() bool {
	return !p.hasAfterCursor() && p.cursor.Before != nil
}

func (p *Paginator) getCursorQuery(op string) string {
	qs := make([]string, len(p.sqlKeys))
	composite := ""
	for i, sqlKey := range p.sqlKeys {
		qs[i] = fmt.Sprintf("%s%s %s ?", composite, sqlKey, op)
		composite = fmt.Sprintf("%s%s = ? AND ", composite, sqlKey)
	}
	return strings.Join(qs, " OR ")
}

func (p *Paginator) getCursorQueryArgs(fields []interface{}) (args []interface{}) {
	for i := 1; i <= len(fields); i++ {
		args = append(args, fields[:i]...)
	}
	return
}

func (p *Paginator) getOperator() string {
	if (p.hasAfterCursor() && p.order == ASC) ||
		(p.hasBeforeCursor() && p.order == DESC) {
		return ">"
	}
	return "<"
}

func (p *Paginator) getOrder() string {
	order := p.order
	if p.hasBeforeCursor() {
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
	if p.hasBeforeCursor() {
		elems.Set(reverse(elems))
	}
	if p.hasBeforeCursor() || hasMore {
		cursor := p.Encode(elems.Index(elems.Len() - 1))
		p.next.After = &cursor
	}
	if p.hasAfterCursor() || (hasMore && p.hasBeforeCursor()) {
		cursor := p.Encode(elems.Index(0))
		p.next.Before = &cursor
	}
	return
}

func (p *Paginator) marshalJSON(value interface{}) []byte {
	rv, ok := value.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(value)
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	fields := make([]interface{}, len(p.keys))
	for i, key := range p.keys {
		fields[i] = rv.FieldByName(key).Interface()
	}
	// @TODO: return proper error
	b, _ := json.Marshal(fields)
	return b
}

func (p *Paginator) unmarshalJSON(bytes []byte) []interface{} {
	var fields []interface{}
	err := json.Unmarshal(bytes, &fields)
	// @TODO: return proper error
	if err != nil {
		return nil
	}
	return p.castJSONFields(fields)
}

func (p *Paginator) castJSONFields(fields []interface{}) []interface{} {
	result := make([]interface{}, len(fields))
	for i, field := range fields {
		kind := p.keyKinds[i]
		switch f := field.(type) {
		case bool:
			bv, err := castJSONBool(f, kind)
			if err != nil {
				return nil
			}
			result[i] = bv
		case float64:
			fv, err := castJSONFloat(f, kind)
			if err != nil {
				return nil
			}
			result[i] = fv
		case string:
			sv, err := castJSONString(f, kind)
			if err != nil {
				return nil
			}
			result[i] = sv
		default:
			// invalid field
			return nil
		}
	}
	return result
}

var (
	errInvalidFieldType = errors.New("invalid field type")
)

func castJSONBool(value bool, kind kind) (interface{}, error) {
	if kind != kindBool {
		return nil, errInvalidFieldType
	}
	return value, nil
}

func castJSONFloat(value float64, kind kind) (interface{}, error) {
	switch kind {
	case kindInt:
		return int(value), nil
	case kindUint:
		return uint(value), nil
	case kindFloat:
		return value, nil
	}
	return nil, errInvalidFieldType
}

func castJSONString(value string, kind kind) (interface{}, error) {
	if kind != kindString && kind != kindTime {
		return nil, errInvalidFieldType
	}
	if kind == kindString {
		return value, nil
	}
	tv, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, errInvalidFieldType
	}
	return tv, nil
}
