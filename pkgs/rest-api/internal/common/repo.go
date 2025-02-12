package common

import "fmt"

type (
	ErrNotFound struct {
		Err error
	}
	ErrTooManyRows struct {
		Err error
	}
	ErrDuplicateKey struct {
		Err error
		Key string
	}
	ErrConstrain struct {
		Err        error
		Constraint string
	}

	ErrUnexpected struct {
		Err error
	}
)

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("not found: %v", e.Err)
}
func (e *ErrNotFound) Unwrap() error { return e.Err }

func (e *ErrTooManyRows) Error() string {
	return fmt.Sprintf("too many rows: %v", e.Err)
}
func (e *ErrTooManyRows) Unwrap() error { return e.Err }

func (e *ErrUnexpected) Error() string {
	return fmt.Sprintf("too many rows: %v", e.Err)
}
func (e *ErrUnexpected) Unwrap() error { return e.Err }

func (e *ErrDuplicateKey) Error() string {
	return fmt.Sprintf("duplicate key violation for %s: %v", e.Key, e.Err)
}
func (e *ErrDuplicateKey) Unwrap() error { return e.Err }

func (e *ErrConstrain) Error() string {
	return fmt.Sprintf("constraint violation for %s: %v", e.Constraint, e.Err)
}
func (e *ErrConstrain) Unwrap() error { return e.Err }
