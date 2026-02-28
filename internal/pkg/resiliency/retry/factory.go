package retry

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RetrierFactory struct {
	Telemetry ports.TelemetryProvider
}

func NewRetrierFactory(telemetry ports.TelemetryProvider) *RetrierFactory {
	return &RetrierFactory{
		Telemetry: telemetry,
	}
}

func (f *RetrierFactory) NewExponentialBackoffRetrier(serviceName string, timeout time.Duration) ports.RetrierInterface {
	cfg := NewRetryConfig(timeout)
	opts := &RetryOptions{
		Config:      cfg,
		Strategy:    NewExponentialBackoff(cfg),
		Telemetry:   f.Telemetry,
		ServiceName: serviceName,
	}

	return NewRetrier(opts)
}

var _ ports.RetrierFactoryInterface = (*RetrierFactory)(nil)
