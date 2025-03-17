package common

import "fmt"

type (
	NotFoundError struct {
		Err error
	}
	DuplicateError struct {
		Err error
		Key string
	}
	ConstrainError struct {
		Err        error
		Constraint string
	}
	UnexpectedError struct {
		Err error
	}
)

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %v", e.Err)
}
func (e *NotFoundError) Unwrap() error { return e.Err }

func (e *UnexpectedError) Error() string {
	return fmt.Sprintf("UnexpectedError: %v", e.Err)
}
func (e *UnexpectedError) Unwrap() error { return e.Err }

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("duplicate key for %s: %v", e.Key, e.Err)
}
func (e *DuplicateError) Unwrap() error { return e.Err }

func (e *ConstrainError) Error() string {
	return fmt.Sprintf("ConstrainError: %v", e.Err)
}
func (e *ConstrainError) Unwrap() error { return e.Err }
