package telemetry

import (
	"strconv"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
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
	circuitBreakerState  *prometheus.GaugeVec   // Labels: method
	circuitBreakerErrors *prometheus.CounterVec // Labels: method, error_type
	requestTimeouts      *prometheus.CounterVec // Labels: method
	rateLimitDrops       *prometheus.CounterVec // Labels: method

	// --- Auth ---
	authFailures *prometheus.CounterVec // Labels: method, reason

	retriesCounter        *prometheus.CounterVec   // Labels: attempt
	retriesDelayHistogram *prometheus.HistogramVec // Labels: attempt
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
		}, []string{"method"}),

		circuitBreakerErrors: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "resiliency_circuit_breaker_errors_total",
			Help:      "Number of times the circuit breaker errors occurred.",
		}, []string{"method", "error_type"}),

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
		retriesCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "resiliency_retries_total",
				Help: "Total number of retry attempts categorized by attempt number",
			},
			[]string{"attempt"},
		),
		// Histogram to track the distribution of sleep/delay times
		retriesDelayHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "resiliency_retry_delay_seconds",
				Help:    "The duration of backoff delays between retries",
				Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}, // Seconds
			},
			[]string{"attempt"},
		),
	}

	return prometheusMetrics
}

func (m *Metrics) GRPCRequestCount(method string, statusCode string) {
	m.grpcRequestCount.WithLabelValues(method, statusCode).Inc()
}

func (m *Metrics) GRPCLatency(method string, durationSeconds float64) {
	m.grpcLatency.WithLabelValues(method).Observe(durationSeconds)
}

func (m *Metrics) CircuitBreakerState(method string, state int) {
	m.circuitBreakerState.WithLabelValues(method).Set(float64(state))
}

func (m *Metrics) CircuitBreakerError(method string, errorType string) {
	m.circuitBreakerErrors.WithLabelValues(method, errorType).Inc()
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

func (m *Metrics) ObserveRetry(attempt int, err error, delay time.Duration) {
	// This method can be used to track retry attempts and their characteristics.
	// For example, you could create additional Prometheus metrics here to count retries or observe delay distributions.
	// This is a simple example that just logs the retry attempt. You can expand this as needed.
	m.grpcRequestCount.WithLabelValues("retry_attempt", strconv.Itoa(attempt)).Inc()
}

var _ MetricsInterface = (*Metrics)(nil)
var _ ports.RetryObserver = (*Metrics)(nil)