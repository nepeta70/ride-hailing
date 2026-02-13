package retry

import (
	"context"
)

type RetrierWithResult[T any] struct {
	baseRetrier
}

func NewRetrierWithResult[T any](opts *RetryOptions) *RetrierWithResult[T] {
	return &RetrierWithResult[T]{
		baseRetrier: baseRetrier{
			config: opts.Config,
		},
	}
}

func (r *RetrierWithResult[T]) DoWithResult(ctx context.Context, op func() (T, error)) (T, error) {
	return r.DoWithResult(ctx, op)
}
