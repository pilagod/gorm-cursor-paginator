package paginator

// pq stands for paging query
type pq struct {
	After  *string
	Before *string
	Limit  *int
	Order  *Order
}

func pqString(str string) *string {
	return &str
}

func pqLimit(limit int) *int {
	return &limit
}

func pqOrder(order Order) *Order {
	return &order
}

func newPaginator(q pq) *Paginator {
	p := New()
	if q.After != nil {
		p.SetAfterCursor(*q.After)
	}
	if q.Before != nil {
		p.SetBeforeCursor(*q.Before)
	}
	if q.Limit != nil {
		p.SetLimit(*q.Limit)
	}
	if q.Order != nil {
		p.SetOrder(*q.Order)
	}
	return p
}
