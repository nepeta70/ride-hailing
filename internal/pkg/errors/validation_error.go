package errors

import "fmt"

type ValidationError struct {
	Code    string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

var ErrInvalidData = &ValidationError{Code: "VALIDATION_ERROR", Message: "invalid data"}

func NewValidationError(cause error) error {
	return &ValidationError{
		Code:    ErrInvalidData.Code,
		Message: ErrInvalidData.Message,
		Err:     BusinessError(cause),
	}
}

func NewValidationErrorf(format string, a ...any) error {
	return NewValidationError(fmt.Errorf(format, a...))
}

func (e *ValidationError) Is(target error) bool {
	t, ok := target.(*ValidationError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}
