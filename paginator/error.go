package paginator

import "errors"

// Errors for paginator
var (
	ErrInvalidLimit = errors.New("limit should be greater than or equal to 0")
	ErrInvalidOrder = errors.New("order should be ASC or DESC")
)
