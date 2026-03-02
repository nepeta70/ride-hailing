// internal/pkg/retry/retryer.go
package retry

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type Retrier struct {
	baseRetrier
}

func NewRetrier(opts *RetryOptions) *Retrier {
	return &Retrier{
		baseRetrier: baseRetrier{
			config:    opts.Config,
			strategy:  opts.Strategy,
			telemetry: opts.Telemetry,
			serviceName: opts.ServiceName,
		},
	}
}

func (r *Retrier) Do(ctx context.Context, op func(ctx context.Context) error) error {
	_, err := r.DoWithResult(ctx, func(ctx context.Context) (any, error) {
		return nil, op(ctx)
	})
	return err
}

var _ ports.RetrierInterface = (*Retrier)(nil)
