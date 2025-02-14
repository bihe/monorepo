package shared

import (
	"errors"
)

// --------------------------------------------------------------------------
//   Error type: NotFoundError
// --------------------------------------------------------------------------

// ErrNotFound creates a new NotFoundError instance
func ErrNotFound(msg string) *NotFoundError {
	return &NotFoundError{
		Err: errors.New(msg),
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

// --------------------------------------------------------------------------
//   Error type: ValidationError
// --------------------------------------------------------------------------

// ErrValidation creates a new ValidationError instance
func ErrValidation(msg string) *ValidationError {
	return &ValidationError{
		Err: errors.New(msg),
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

// --------------------------------------------------------------------------
//   Error type: SecurityError
// --------------------------------------------------------------------------

// ErrSecurity creates a new SecurityError instance
func ErrSecurity(msg string) *SecurityError {
	return &SecurityError{
		Err: errors.New(msg),
	}
}

// SecurityError is used when there is an access-problem caused by wrong/missing permissions
type SecurityError struct {
	Err error
}

// Error is the string representation of the error
func (e *SecurityError) Error() string {
	return e.Err.Error()
}
