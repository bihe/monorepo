package shared

import (
	"fmt"
)

// --------------------------------------------------------------------------
// Error defintions
// --------------------------------------------------------------------------

// ErrNotFound
// --------------------------------------------------------------------------

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

// Error is the string representation of the error
func (e *NotFoundError) Error() string {
	return e.Err.Error()
}

// ErrValidation
// --------------------------------------------------------------------------

// ErrValidation creates a typed error with the given message
func ErrValidation(msg string) *ValidationError {
	return &ValidationError{
		Err: fmt.Errorf(msg),
	}
}

// ValidationError is used to indicate that missing or invalid parameters are supplied
type ValidationError struct {
	Err error
}

// Error is the string representation of the error
func (e *ValidationError) Error() string {
	return e.Err.Error()
}
