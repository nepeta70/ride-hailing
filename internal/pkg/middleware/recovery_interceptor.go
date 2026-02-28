package middleware

import (
	"context"
	"runtime/debug"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ServerInterceptor manages observability and resiliency for gRPC streams.
type ServerInterceptor struct {
	telemetry ports.TelemetryProvider
}

// NewRecoveryInterceptor initializes a new interceptor with necessary dependencies.
func NewRecoveryInterceptor(telemetry ports.TelemetryProvider) *ServerInterceptor {
	return &ServerInterceptor{
		telemetry: telemetry,
	}
}

func (i *ServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	tr := i.telemetry.Tracer()
	pr := i.telemetry.Propagator()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		defer func() {
			if r := recover(); r != nil {
				i.telemetry.Logger().ErrorContext(ctx, "gRPC panic recovered",
					"error", r,
					"method", info.FullMethod,
					"stack", string(debug.Stack()))
			}
		}()

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = pr.Extract(ctx, propagation.HeaderCarrier(md))
		}

		ctx, span := tr.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		return handler(ctx, req)
	}
}
