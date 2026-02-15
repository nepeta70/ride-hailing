package middleware

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gRPC Interceptor using our decoupled Logger interface
func UnaryServerLogging(l ports.Logger, metrics telemetry.MetricsInterface) grpc.UnaryServerInterceptor {
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

			st, _ := status.FromError(err)
			metrics.GRPCRequestCount(info.FullMethod, st.Code().String())
			duration := time.Since(start).Seconds()
			metrics.GRPCLatency(info.FullMethod, duration)
			l.Info("gRPC request processed",
				"method", info.FullMethod,
				"status", uint32(st.Code()),
				"duration_s", duration,
			)
		}()

		// Execute the actual RPC handler
		resp, err = handler(ctx, req)
		return resp, err
	}
}
