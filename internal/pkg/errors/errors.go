package errors

import (
	"errors"
	"fmt"
)

// ErrorCategory defines the type of error for handling strategy
type ErrorCategory int

const (
	// Transient errors are temporary and should be retried with backoff
	Transient ErrorCategory = iota
	// Permanent errors are unrecoverable and should not be retried
	Permanent
	// Business errors are domain-specific validation or state errors
	Business
)

// CategorizedError wraps an error with its category
type CategorizedError struct {
	Category ErrorCategory
	Err      error
}

func (e *CategorizedError) Error() string {
	return e.Err.Error()
}

func (e *CategorizedError) Unwrap() error {
	return e.Err
}

// NewTransientError creates a new transient error
func NewTransientError(msg string) error {
	return &CategorizedError{
		Category: Transient,
		Err:      errors.New(msg),
	}
}

// NewTransientErrorf creates a new transient error with formatting
func NewTransientErrorf(format string, args ...any) error {
	return &CategorizedError{
		Category: Transient,
		Err:      fmt.Errorf(format, args...),
	}
}

// NewPermanentError creates a new permanent error
func NewPermanentError(msg string) error {
	return &CategorizedError{
		Category: Permanent,
		Err:      errors.New(msg),
	}
}

// NewPermanentErrorf creates a new permanent error with formatting
func NewPermanentErrorf(format string, args ...any) error {
	return &CategorizedError{
		Category: Permanent,
		Err:      fmt.Errorf(format, args...),
	}
}

// NewBusinessError creates a new business logic error
func NewBusinessError(msg string) error {
	return BusinessError(errors.New(msg))
}

// NewBusinessErrorf creates a new business logic error with formatting
func NewBusinessErrorf(format string, args ...any) error {
	return BusinessError(fmt.Errorf(format, args...))
}

func BusinessError(err error) error {
	if err == nil {
		return nil
	}
	return &CategorizedError{
		Category: Business,
		Err:      err,
	}
}

// IsTransient returns true if the error should be retried
func IsTransient(err error) bool {
	var catErr *CategorizedError
	if errors.As(err, &catErr) {
		return catErr.Category == Transient
	}
	return false
}

// IsPermanent returns true if the error should not be retried
func IsPermanent(err error) bool {
	var catErr *CategorizedError
	if errors.As(err, &catErr) {
		return catErr.Category == Permanent
	}
	return false
}

// IsBusiness returns true if the error is a business logic error
func IsBusiness(err error) bool {
	var catErr *CategorizedError
	if errors.As(err, &catErr) {
		return catErr.Category == Business
	}
	return false
}

func NewErrJSONMarshal(err error) error {
	return NewPermanentErrorf("failed to marshal to json: %w", err)
}

func NewErrJSONUnmarshal(err error) error {
	return NewPermanentErrorf("failed to unmarshal from json: %w", err)
}

func NewErrNotFound(msg string) error {
	return NewBusinessErrorf("not found: %s", msg)
}

// IsNotFound checks if the error message contains 'not found'.
func IsNotFound(err error) bool {
	return err != nil && (len(err.Error()) > 9 && err.Error()[:9] == "not found")
}

var (
	ErrContextError  = NewTransientError("context cancelled or deadline exceeded")
	ErrInternalError = NewTransientError("internal error")
)
