package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsInterface interface {
	// --- RPC Metrics ---
	GRPCRequestCount(method string, statusCode string)
	GRPCLatency(method string, durationSeconds float64)

	// --- Resiliency & Stability ---
	CircuitBreakerState(serviceName string, state int)
	CircuitBreakerError(serviceName string, errorType string)
	RequestTimeout(method string)
	RateLimitDrop(method string)

	// --- Auth ---
	AuthFailure(method string, reason string)
}

type Metrics struct {
	// --- RPC Metrics ---
	grpcRequestCount *prometheus.CounterVec   // Labels: method, status_code
	grpcLatency      *prometheus.HistogramVec // Labels: method

	// --- Resiliency & Stability ---
	circuitBreakerState  *prometheus.GaugeVec   // Labels: service_name
	circuitBreakerErrors *prometheus.CounterVec // Labels: service_name, error_type
	requestTimeouts      *prometheus.CounterVec // Labels: method
	rateLimitDrops       *prometheus.CounterVec // Labels: method

	// --- Auth ---
	authFailures *prometheus.CounterVec // Labels: method, reason
}

func NewMetrics(namespace string, subSystem string, reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)

	prometheusMetrics := &Metrics{
		grpcRequestCount: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests.",
		}, []string{"method", "status_code"}),

		grpcLatency: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "grpc_latency_seconds",
			Help:      "Latency of gRPC requests.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method"}),

		circuitBreakerState: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "resiliency_circuit_breaker_state",
			Help:      "State of the circuit breaker (0=Closed, 1=Open, 2=HalfOpen)",
		}, []string{"service_name"}),

		circuitBreakerErrors: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "resiliency_circuit_breaker_errors_total",
			Help:      "Number of times the circuit breaker errors occurred.",
		}, []string{"service_name", "error_type"}),

		requestTimeouts: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "resiliency_timeout_total",
			Help:      "Total requests that exceeded the defined timeout",
		}, []string{"method"}),

		rateLimitDrops: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "resiliency_rate_limit_dropped_total",
			Help:      "Total requests dropped due to rate limiting",
		}, []string{"method"}),

		authFailures: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "security_auth_failures_total",
			Help:      "Total number of failed authentication or authorization attempts",
		}, []string{"method", "reason"}),
	}

	return prometheusMetrics
}

func (m *Metrics) GRPCRequestCount(method string, statusCode string) {
	m.grpcRequestCount.WithLabelValues(method, statusCode).Inc()
}

func (m *Metrics) GRPCLatency(method string, durationSeconds float64) {
	m.grpcLatency.WithLabelValues(method).Observe(durationSeconds)
}

func (m *Metrics) CircuitBreakerState(serviceName string, state int) {
	m.circuitBreakerState.WithLabelValues(serviceName).Set(float64(state))
}

func (m *Metrics) CircuitBreakerError(serviceName string, errorType string) {
	m.circuitBreakerErrors.WithLabelValues(serviceName, errorType).Inc()
}

func (m *Metrics) RequestTimeout(method string) {
	m.requestTimeouts.WithLabelValues(method).Inc()
}

func (m *Metrics) RateLimitDrop(method string) {
	m.rateLimitDrops.WithLabelValues(method).Inc()
}

func (m *Metrics) AuthFailure(method string, reason string) {
	m.authFailures.WithLabelValues(method, reason).Inc()
}

var _ MetricsInterface = (*Metrics)(nil)
