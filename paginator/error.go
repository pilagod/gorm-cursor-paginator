package paginator

import "errors"

// Errors for paginator
var (
	ErrEmptyRule    = errors.New("rule cannot be empty")
	ErrInvalidLimit = errors.New("limit should be greater than 0")
	ErrInvalidModel = errors.New("model fields should match rules or keys specified for paginator")
	ErrInvalidOrder = errors.New("order should be ASC or DESC")
)
