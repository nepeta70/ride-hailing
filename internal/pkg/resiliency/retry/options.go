package retry

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RetryOptions struct {
	Config   *RetryConfig
	Strategy RetryStrategy
	Logger   ports.Logger
}

func (o *RetryOptions) Validate() error {
	if o.Strategy == nil {
		return errors.NewValidationErrorf("Strategy cannot be nil")
	}
	if o.Logger == nil {
		return errors.NewValidationErrorf("Logger cannot be nil")
	}
	return nil
}
