package middleware

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/circuitbreaker"
	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"google.golang.org/grpc"
)

func UnaryCircuitBreaker(cb *circuitbreaker.CircuitBreaker, metrics telemetry.MetricsInterface) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		var resp any

		// We use your Execute method to wrap the handler
		err := cb.Execute(ctx, func() error {
			var err error
			resp, err = handler(ctx, req)
			return err
		})

		// Check if the error returned came from the CB logic itself
		if err != nil {
			switch err {
			case circuitbreaker.ErrCircuitOpen:
				metrics.CircuitBreakerError(info.FullMethod, "circuit_open")

			case circuitbreaker.ErrTooManyRequests:
				metrics.CircuitBreakerError(info.FullMethod, "half_open_limit")

			case circuitbreaker.ErrPanicRecovered:
				metrics.CircuitBreakerError(info.FullMethod, "panic_recovered")

			default:
			}
		}

		return resp, err
	}
}
