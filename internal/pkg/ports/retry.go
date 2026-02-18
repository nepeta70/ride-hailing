package ports

import (
	"context"
	"time"
)

type RetryObserver interface {
	ObserveRetry(attempt int, err error, delay time.Duration)
}

type RetrierInterface interface {
	Do(ctx context.Context, op func() error) error
	DoWithResult(ctx context.Context, op func() (any, error)) (any, error)
}

type RetrierFactoryInterface interface {
	NewExponentialBackoffRetrier(timeout time.Duration) RetrierInterface
}
