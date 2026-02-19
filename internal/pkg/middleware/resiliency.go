package middleware

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/circuitbreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

// ResiliencyInterceptor handles rate limiting and circuit breaking.
type ResiliencyInterceptor struct {
	circuitBreaker *circuitbreaker.CircuitBreaker
	metrics        ports.Metrics
	limiter        *rate.Limiter
}

// NewResiliencyInterceptor initializes the interceptor with resiliency dependencies.
func NewResiliencyInterceptor(rateLimit float64, rateBurst int, m ports.Metrics) (*ResiliencyInterceptor, error) {
	limiter := rate.NewLimiter(rate.Limit(rateLimit), rateBurst)
	cb, err := circuitbreaker.NewCircuitBreaker(circuitbreaker.DefaultConfig())
	if err != nil {
		return nil, err
	}
	return &ResiliencyInterceptor{
		circuitBreaker: cb,
		metrics:        m,
		limiter:        limiter,
	}, nil
}

// Unary provides the gRPC interceptor logic for rate limiting and circuit breaking.
func (i *ResiliencyInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		// 1. Rate Limiting
		if !i.limiter.Allow() {
			i.metrics.RateLimitDrop(info.FullMethod)
			return nil, errResourceExhausted
		}

		var resp any

		// 2. Circuit Breaker Execution
		// Note: The closure here is necessary to bridge the gRPC handler to the CB Execute signature.
		err := i.circuitBreaker.Execute(ctx, func() error {
			var err error
			resp, err = handler(ctx, req)
			return err
		})

		// 3. CB Error Mapping & Metrics
		if err != nil {
			switch err {
			case circuitbreaker.ErrCircuitOpen:
				i.metrics.CircuitBreakerError(info.FullMethod, "circuit_open")

			case circuitbreaker.ErrTooManyRequests:
				i.metrics.CircuitBreakerError(info.FullMethod, "half_open_limit")

			case circuitbreaker.ErrPanicRecovered:
				i.metrics.CircuitBreakerError(info.FullMethod, "panic_recovered")
			}
		}

		return resp, err
	}
}
