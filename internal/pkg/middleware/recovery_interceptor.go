package middleware

import (
	"context"
	"runtime/debug"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"google.golang.org/grpc"
)

// ServerInterceptor manages observability and resiliency for gRPC streams.
type ServerInterceptor struct {
	logger ports.Logger
}

// NewRecoveryInterceptor initializes a new interceptor with necessary dependencies.
func NewRecoveryInterceptor(logger ports.Logger) *ServerInterceptor {
	return &ServerInterceptor{
		logger: logger,
	}
}

func (i *ServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		defer func() {
			if r := recover(); r != nil {
				i.logger.ErrorContext(ctx, "gRPC panic recovered",
					"error", r,
					"method", info.FullMethod,
					"stack", string(debug.Stack()))
			}
		}()
		return handler(ctx, req)
	}
}
