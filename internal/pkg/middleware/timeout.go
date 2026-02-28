package middleware

import (
	"context"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
)

// TimeoutInterceptor manages request deadlines and timeout telemetry.
type TimeoutInterceptor struct {
	timeout   time.Duration
	telemetry ports.TelemetryProvider
}

// NewTimeoutInterceptor initializes the interceptor struct with dependencies.
func NewTimeoutInterceptor(timeout time.Duration, telemetry ports.TelemetryProvider) *TimeoutInterceptor {
	return &TimeoutInterceptor{
		timeout:   timeout,
		telemetry: telemetry,
	}
}

// Unary provides the gRPC interceptor logic for forcing request timeouts via goroutines.
func (i *TimeoutInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Create a new context with the specified timeout
		ctx, cancel := context.WithTimeout(ctx, i.timeout)
		defer cancel()

		// Channels to capture the result
		respChan := make(chan any, 1)
		errChan := make(chan error, 1)

		go func() {
			resp, err := handler(ctx, req)
			respChan <- resp
			errChan <- err
		}()

		select {
		case <-ctx.Done():
			i.telemetry.Metrics().RequestTimeout(info.FullMethod)
			return nil, errDeadlineExceeded
		case err := <-errChan:
			return <-respChan, err
		}
	}
}
