package middleware

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/circuitbreaker"
	"google.golang.org/grpc"
)

func UnaryCircuitBreaker(cb *circuitbreaker.CircuitBreaker) grpc.UnaryServerInterceptor {
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

		// If Execute returns an error, it could be from the handler
		// OR a circuit breaker error (ErrCircuitOpen / ErrTooManyRequests)
		return resp, err
	}
}
