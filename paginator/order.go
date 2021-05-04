package paginator

// Order type for order
type Order string

// Orders
const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

// Flip flips order
func (o *Order) Flip() Order {
	if *o == ASC {
		return DESC
	}
	return ASC
}

// Validate validates order
func (o *Order) Validate(allowEmpty bool) error {
	if *o == ASC || *o == DESC {
		return nil
	}
	if allowEmpty && *o == "" {
		return nil
	}
	return ErrInvalidOrder
}
