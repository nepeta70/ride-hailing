package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// MockMetrics is a thread-safe mock for MetricsInterface
type MockMetrics struct {
	mu sync.Mutex

	// Maps to track call counts and labels for verification
	Calls map[string]int
	Args  map[string][]any
}

func NewMockMetrics() *MockMetrics {
	return &MockMetrics{
		Calls: make(map[string]int),
		Args:  make(map[string][]any),
	}
}

func (m *MockMetrics) track(name string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls[name]++
	m.Args[name] = args
}

// --- Interface Implementation ---
func (m *MockMetrics) CircuitBreakerState(serviceName string, state int) {
	m.track("CircuitBreakerState", serviceName, state)
}

func (m *MockMetrics) CircuitBreakerError(serviceName string, errorType string) {
	m.track("CircuitBreakerError", serviceName, errorType)
}

func (m *MockMetrics) RequestTimeout(method string) {
	m.track("RequestTimeout", method)
}

func (m *MockMetrics) RateLimitDrop(method string) {
	m.track("RateLimitDrop", method)
}

func (m *MockMetrics) AuthFailure(method string, reason string) {
	m.track("AuthFailure", method, reason)
}

func (m *MockMetrics) ObserveRetry(service string, attempt int, err error, delay time.Duration) {
	m.track("ObserveRetry", service, attempt, err, delay)
}

func (m *MockMetrics) ValidationFailure(method string, reason string) {
	m.track("ValidationFailure", method, reason)
}

func (m *MockMetrics) DependencyFailure(dependency string, operation string, errorType string) {
	m.track("DependencyFailure", dependency, operation, errorType)
}

var _ ports.Metrics = (*MockMetrics)(nil)

type MockTelemetryProvider struct {
	metrics *MockMetrics
	logger  *MockLogger
}

func NewMockTelemetryProvider() *MockTelemetryProvider {
	return &MockTelemetryProvider{
		metrics: NewMockMetrics(),
		logger:  NewMockLogger(),
	}
}

func (m *MockTelemetryProvider) Metrics() ports.Metrics {
	return m.metrics
}

func (m *MockTelemetryProvider) Logger() ports.Logger {
	return m.logger
}

func (m *MockTelemetryProvider) Tracer() trace.Tracer {
	return nil
}

func (m *MockTelemetryProvider) Propagator() propagation.TextMapPropagator {
	return nil
}

func (m *MockTelemetryProvider) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MockTelemetryProvider) LogEntries() []string {
	return m.logger.Entries
}

func (m *MockTelemetryProvider) MetricsCalls() map[string]int {
	return m.metrics.Calls
}

func (m *MockTelemetryProvider) MetricsArgs() map[string][]any {
	return m.metrics.Args
}

var _ ports.TelemetryProvider = (*MockTelemetryProvider)(nil)
