package retry

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RetrierWithResult[T any] struct {
	baseRetrier
}

func NewRetrierWithResult[T any](cfg RetryConfig, strategy RetryStrategy, logger ports.Logger) *RetrierWithResult[T] {
	return &RetrierWithResult[T]{

		baseRetrier: baseRetrier{
			config: cfg,
		},
	}
}

func (r *RetrierWithResult[T]) DoWithResult(ctx context.Context, op func() (T, error)) (T, error) {
	return r.DoWithResult(ctx, op)
}
