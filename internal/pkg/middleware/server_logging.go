package middleware

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// ServerInterceptor manages observability and resiliency for gRPC streams.
type ServerInterceptor struct {
	logger  ports.Logger
	metrics ports.Metrics
}

// NewServerInterceptor initializes a new interceptor with necessary dependencies.
func NewServerInterceptor(l ports.Logger, m ports.Metrics) *ServerInterceptor {
	return &ServerInterceptor{
		logger:  l,
		metrics: m,
	}
}

// Unary provides logging, metrics collection, and panic recovery.
func (i *ServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		defer func() {
			// 1. Recovery Logic
			if r := recover(); r != nil {
				i.logger.Error("gRPC panic recovered",
					"error", r,
					"method", info.FullMethod,
					"stack", string(debug.Stack()),
				)
				err = errInternal
			}

			// 2. Telemetry & Observability
			st, _ := status.FromError(err)
			codeStr := st.Code().String()
			duration := time.Since(start).Seconds()

			i.metrics.GRPCRequestCount(info.FullMethod, codeStr)
			i.metrics.GRPCLatency(info.FullMethod, duration)

			i.logger.Info("gRPC request processed",
				"method", info.FullMethod,
				"status", uint32(st.Code()),
				"duration_s", duration,
			)
		}()

		return handler(ctx, req)
	}
}
