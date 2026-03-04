package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, span := i.telemetry.Tracer().Start(ctx, "Middleware.RecoveryInterceptor", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		defer func() {
			if r := recover(); r != nil {
				span.SetStatus(otelcodes.Error, "Panic recovered in gRPC handler")
				span.RecordError(fmt.Errorf("%v", r))
				i.telemetry.Logger().ErrorContext(ctx, "gRPC panic recovered",
					"error", r,
					"method", info.FullMethod,
					"stack", string(debug.Stack()))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		if err == nil {
			span.SetStatus(otelcodes.Ok, "gRPC handler executed successfully")
		}
		return handler(ctx, req)
	}
}
