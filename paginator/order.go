package paginator

// Order type for order
type Order string

// Orders
const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

func (o *Order) Flip() Order {
	if *o == ASC {
		return DESC
	}
	return ASC
}
