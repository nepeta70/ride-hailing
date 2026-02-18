package retry

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RetrierFactory struct {
	Logger  ports.Logger
	Metrics ports.RetryObserver
}

func NewRetrierFactory(logger ports.Logger, metrics ports.RetryObserver) *RetrierFactory {
	return &RetrierFactory{
		Logger:  logger,
		Metrics: metrics,
	}
}

func (f *RetrierFactory) NewExponentialBackoffRetrier(timeout time.Duration) ports.RetrierInterface {
	cfg := NewRetryConfig(timeout)
	opts := &RetryOptions{
		Config:   cfg,
		Strategy: newExponentialBackoff(cfg),
		Logger:   f.Logger,
		Metrics:  f.Metrics,
	}

	return NewRetrier(opts)
}
