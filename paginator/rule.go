package paginator

// Rule for paginator
type Rule struct {
	Key     string
	Order   Order
	SQLRepr string
}

func (r *Rule) Validate() error {
	if err := r.Order.Validate(true); err != nil {
		return err
	}
	return nil
}
