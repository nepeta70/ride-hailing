package mocks

import (
	"sync"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
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
