package middleware

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gRPC Interceptor using our decoupled Logger interface
func UnaryServerLogging(l ports.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		// Panic Recovery
		defer func() {
			if r := recover(); r != nil {
				l.Error("gRPC panic recovered",
					"error", r,
					"method", info.FullMethod,
					"stack", string(debug.Stack()),
				)
				// Convert panic to a proper gRPC Internal error
				err = status.Errorf(codes.Internal, "internal server error")
			}

			duration := time.Since(start)
			st, _ := status.FromError(err)

			l.Info("gRPC request processed",
				"method", info.FullMethod,
				"status", st.Code().String(),
				"duration", duration,
			)
		}()

		// Execute the actual RPC handler
		resp, err = handler(ctx, req)
		return resp, err
	}
}
