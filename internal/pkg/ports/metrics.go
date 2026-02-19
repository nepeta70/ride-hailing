package ports

type Metrics interface {
	GRPCMetrics
	RetryObserver
	CircuitBreakerMetrics
	RequestTimeout(method string)
	RateLimitDrop(method string)

	// --- Auth ---
	AuthFailure(method string, reason string)
	ValidationFailure(method string, reason string)
}

type GRPCMetrics interface {
	GRPCRequestCount(method string, statusCode string)
	GRPCLatency(method string, durationSeconds float64)
}

type CircuitBreakerMetrics interface {
	CircuitBreakerState(serviceName string, state int)
	CircuitBreakerError(serviceName string, errorType string)
}
