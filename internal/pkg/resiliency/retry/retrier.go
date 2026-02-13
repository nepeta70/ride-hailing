// internal/pkg/retry/retryer.go
package retry

import (
	"context"
)

type Retrier struct {
	baseRetrier
}

func NewRetrier(opts *RetryOptions) *Retrier {
	return &Retrier{
		baseRetrier: baseRetrier{
			config:   opts.Config,
			strategy: opts.Strategy,
			logger:   opts.Logger,
		},
	}
}

func (r *Retrier) Do(ctx context.Context, op func() error) error {
	_, err := r.DoWithResult(ctx, func() (any, error) {
		return nil, op()
	})
	return err
}
