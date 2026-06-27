package ports

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	tracer "go.opentelemetry.io/otel/trace"
)

type Metrics interface {
	RetryObserver
	CircuitBreakerMetrics
	RequestTimeout(method string)
	RateLimitDrop(method string)

	// --- Auth ---
	AuthFailure(method string, reason string)
	ValidationFailure(method string, reason string)
	DependencyFailure(dependency string, operation string, errorType string)
}

type CircuitBreakerMetrics interface {
	CircuitBreakerState(serviceName string, state int)
	CircuitBreakerError(serviceName string, errorType string)
}

type TelemetryProvider interface {
	Metrics() Metrics
	Logger() Logger
	Tracer() tracer.Tracer
	Propagator() propagation.TextMapPropagator
	TracerProvider() *trace.TracerProvider
	MeterProvider() *sdkmetric.MeterProvider
	Shutdown(ctx context.Context) error
}
