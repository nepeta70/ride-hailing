package errors

import "fmt"

type ValidationError struct {
	Code    string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

var ErrInvalidBookingData = &ValidationError{Code: "VALIDATION_ERROR", Message: "invalid booking data"}

func NewValidationError(cause error) error {
	err := &ValidationError{
		Code:    ErrInvalidBookingData.Code,
		Message: ErrInvalidBookingData.Message,
		Err:     cause,
	}

	return BusinessError(err)
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
