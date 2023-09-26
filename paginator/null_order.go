package paginator

import "fmt"

// NullOrder type for order for null values
type NullOrder string

// Null orders
const (
	FIRST NullOrder = "FIRST"
	LAST  NullOrder = "LAST"
)

func (o *NullOrder) flip() NullOrder {
	if *o == FIRST {
		return LAST
	}
	return FIRST
}

func (o *NullOrder) validate() error {
	if *o != FIRST && *o != LAST {
		return ErrInvalidOrder
	}
	return nil
}

// NullOrderBuilder builds query string for null value ordering. Note that `rule.NullOrder` is assumed to be valid.
type NullOrderBuilder func(rule Rule, order Order) string

// PostgresNullOrderBuilder is the null order builder for PostgreSQL, Oracle, and SQLite.
func PostgresNullOrderBuilder(rule Rule, order Order) string {
	return fmt.Sprintf("%s %s NULLS %s", rule.SQLRepr, order, rule.NullOrder)
}

// MySQLNullOrderBuilder is the null order builder for MySQL and SQL Server.
func MySQLNullOrderBuilder(rule Rule, order Order) string {
	stmt := fmt.Sprintf("%s %s", rule.SQLRepr, order)
	if rule.NullOrder == FIRST {
		return fmt.Sprintf("%s IS NULL, %s", rule.SQLRepr, stmt)
	}
	return fmt.Sprintf("%s IS NOT NULL, %s", rule.SQLRepr, stmt)
}
