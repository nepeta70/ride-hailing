package retry

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RetryOptions struct {
	Config      *RetryConfig
	Strategy    RetryStrategy
	Logger      ports.Logger
	Metrics     ports.RetryObserver
	ServiceName string
}

func (o *RetryOptions) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("Config cannot be nil")
	}
	if o.Strategy == nil {
		return errors.NewValidationErrorf("Strategy cannot be nil")
	}
	if o.Logger == nil {
		return errors.NewValidationErrorf("Logger cannot be nil")
	}
	if o.Metrics == nil {
		return errors.NewValidationErrorf("Metrics cannot be nil")
	}
	if o.ServiceName == "" {
		return errors.NewValidationErrorf("ServiceName cannot be empty")
	}
	return nil
}
