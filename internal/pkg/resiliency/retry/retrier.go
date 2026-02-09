// internal/pkg/retry/retryer.go
package retry

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type Retrier struct {
	baseRetrier
}

func NewRetrier(cfg RetryConfig, strategy RetryStrategy, logger ports.Logger) *Retrier {
	return &Retrier{
		baseRetrier: baseRetrier{
			config:   cfg,
			strategy: strategy,
			logger:   logger,
		},
	}
}

func (r *Retrier) Do(ctx context.Context, op func() error) error {
	_, err := r.DoWithResult(ctx, func() (any, error) {
		return nil, op()
	})
	return err
}
