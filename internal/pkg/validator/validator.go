package validator

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type Validator struct {
	errs []error
}

func New() *Validator {
	return &Validator{errs: []error{}}
}

func (v *Validator) HasErrors() bool {
	return len(v.errs) > 0
}

func (v *Validator) Errors() error {
	if !v.HasErrors() {
		return nil
	}
	return errors.NewValidationErrorf("validation errors: %s", v.errs)
}

// Field is the entry point for a fluent chain
func (v *Validator) StringField(name string, value string) *StringFieldValidator {
	return &StringFieldValidator{
		parent: v,
		name:   name,
		value:  value,
	}
}

func (v *Validator) FloatField(name string, value float64) *FloatFieldValidator {
	return &FloatFieldValidator{
		parent: v,
		name:   name,
		value:  value,
	}
}

type StringFieldValidator struct {
	parent *Validator
	name   string
	value  string
}

func (fv *StringFieldValidator) Required() *StringFieldValidator {
	if fv.value == "" {
		fv.parent.errs = append(fv.parent.errs, errors.NewValidationErrorf("%s is required", fv.name))
	}

	return fv
}

func (fv *StringFieldValidator) IsUUID() *StringFieldValidator {
	if err := uuid.Validate(fv.value); err != nil {
		fv.parent.errs = append(fv.parent.errs, errors.NewValidationErrorf("%s must be a valid UUID", fv.name))
	}
	return fv
}

type FloatFieldValidator struct {
	parent *Validator
	name   string
	value  float64
}

func (fv *FloatFieldValidator) GreaterThanZero() *FloatFieldValidator {
	return fv.GreaterThan(0)
}

func (fv *FloatFieldValidator) GreaterThan(min float64) *FloatFieldValidator {
	if fv.value <= min {
		fv.parent.errs = append(fv.parent.errs, errors.NewValidationErrorf("%s must be greater than %f", fv.name, min))
	}
	return fv
}
