package retry

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	config RetryConfig
}

func NewExponentialBackoff(cfg RetryConfig) *ExponentialBackoff {
	return &ExponentialBackoff{config: cfg}
}

// NextDelay calculates the next delay with exponential backoff and jitter
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(e.config.InitialDelay)
	if attempt > 1 {
		delay *= math.Pow(e.config.Multiplier, float64(attempt-1))
	}

	// Cap at max delay
	maxDelay := float64(e.config.MaxDelay)
	delay = math.Min(delay, maxDelay)

	// Add jitter: randomize +/- jitterFraction of the delay
	if e.config.JitterFraction > 0 {
		jitterAmount := delay * e.config.JitterFraction
		jitter := (rand.Float64() * 2 * jitterAmount) - jitterAmount
		delay = math.Max(0, delay+jitter)
	}

	return time.Duration(delay)
}

func NewExponentialBackoffRetrierWithTimeout(timeout time.Duration, logger ports.Logger) *Retrier {
	cfg := NewRetryConfig(timeout)
	return NewRetrier(cfg, NewExponentialBackoff(cfg), logger)
}
