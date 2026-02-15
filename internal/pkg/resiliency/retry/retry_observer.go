package retry

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	retriesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resiliency_retries_total",
			Help: "Total number of retry attempts categorized by attempt number",
		},
		[]string{"attempt"},
	)
	// Histogram to track the distribution of sleep/delay times
	retriesDelayHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "resiliency_retry_delay_seconds",
			Help:    "The duration of backoff delays between retries",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}, // Seconds
		},
		[]string{"attempt"},
	)
)

type PromRetryObserver struct{}

func (p *PromRetryObserver) ObserveRetry(attempt int, err error, delay time.Duration) {
	// Convert attempt int to string for Prometheus label
	retriesCounter.WithLabelValues(strconv.Itoa(attempt)).Inc()
	retriesDelayHistogram.WithLabelValues(strconv.Itoa(attempt)).Observe(delay.Seconds())
}
