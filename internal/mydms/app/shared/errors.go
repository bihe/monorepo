package shared

import (
	"fmt"
)

// ErrNotFound creates a typed error with the given message
func ErrNotFound(msg string) *NotFoundError {
	return &NotFoundError{
		Err: fmt.Errorf(msg),
	}
}

// NotFoundError is used to indicate that a elements is not available
type NotFoundError struct {
	Err error
}

func (e *NotFoundError) Error() string {
	return e.Err.Error()
}
