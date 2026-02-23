package ports

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	GetMetrics() Metrics
	GetLogger() Logger
	GetTracer() trace.Tracer
	GetPropagator() propagation.TextMapPropagator
	Shutdown(ctx context.Context) error
}
