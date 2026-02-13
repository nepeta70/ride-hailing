package retry

import "time"

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay before first retry
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Backoff multiplier (e.g., 2.0 for exponential)
	JitterFraction  float64       // Fraction of delay to use for jitter (0.0-1.0)
	RetryableErrors []error       // Specific errors to retry (optional)
}

// DefaultConfig returns a sensible default retry configuration
func DefaultConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:    3,
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       10 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.1,
	}
}

// AggressiveConfig for critical operations that need more retries
func AggressiveConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:    5,
		InitialDelay:   50 * time.Millisecond,
		MaxDelay:       30 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.2,
	}
}

func NewRetryConfig(timeout time.Duration) *RetryConfig {
	return &RetryConfig{
		MaxAttempts:    5,
		InitialDelay:   timeout / 20, // 5% of total time
		MaxDelay:       timeout / 4,  // 25% of total time
		Multiplier:     1.5,          // Slower growth
		JitterFraction: 0.1,
	}
}
