package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryTimeout(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Create a new context with the specified timeout
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Channel to capture the result
		respChan := make(chan any, 1)
		errChan := make(chan error, 1)

		go func() {
			resp, err := handler(ctx, req)
			respChan <- resp
			errChan <- err
		}()

		select {
		case <-ctx.Done():
			return nil, status.Error(codes.DeadlineExceeded, "request timed out")
		case err := <-errChan:
			return <-respChan, err
		}
	}
}
