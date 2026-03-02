package ports

import (
	"context"
	"time"
)

type RetryObserver interface {
	ObserveRetry(service string, attempt int, err error, delay time.Duration)
}

type RetrierInterface interface {
	Do(ctx context.Context, op func(ctx context.Context) error) error
	DoWithResult(ctx context.Context, op func(ctx context.Context) (any, error)) (any, error)
}

type RetrierFactoryInterface interface {
	NewExponentialBackoffRetrier(serviceName string, timeout time.Duration) RetrierInterface
}
