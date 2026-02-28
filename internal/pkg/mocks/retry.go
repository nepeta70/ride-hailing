package mocks

import (
	"context"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type NoOpRetrier struct{}

func (r *NoOpRetrier) Do(ctx context.Context, op func(ctx context.Context) error) error {
	return op(ctx)
}

func (r *NoOpRetrier) DoWithResult(ctx context.Context, op func(ctx context.Context) (any, error)) (any, error) {
	return op(ctx)
}

var _ ports.RetrierInterface = (*NoOpRetrier)(nil)

type NoOpRetrierFactory struct{}

func (f *NoOpRetrierFactory) NewExponentialBackoffRetrier(serviceName string, timeout time.Duration) ports.RetrierInterface {
	return &NoOpRetrier{}
}

var _ ports.RetrierFactoryInterface = (*NoOpRetrierFactory)(nil)
