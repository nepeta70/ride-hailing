package retry

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RetryOptions struct {
	Config      *RetryConfig
	Strategy    RetryStrategy
	Telemetry   ports.TelemetryProvider
	ServiceName string
}

func (o *RetryOptions) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("Config cannot be nil")
	}
	if o.Strategy == nil {
		return errors.NewValidationErrorf("Strategy cannot be nil")
	}
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("Telemetry cannot be nil")
	}
	if o.ServiceName == "" {
		return errors.NewValidationErrorf("ServiceName cannot be empty")
	}
	return nil
}
