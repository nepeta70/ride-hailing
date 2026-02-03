package retry

import (
	"context"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

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
func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       10 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.1,
	}
}

// AggressiveConfig for critical operations that need more retries
func AggressiveConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    5,
		InitialDelay:   50 * time.Millisecond,
		MaxDelay:       30 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.2,
	}
}

// RetryStrategy defines how to handle retries
type RetryStrategy interface {
	ShouldRetry(err error, attempt int) bool
	NextDelay(attempt int) time.Duration
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	Config RetryConfig
}

func NewExponentialBackoff(cfg RetryConfig) *ExponentialBackoff {
	return &ExponentialBackoff{Config: cfg}
}

// ShouldRetry determines if an error is retryable
func (e *ExponentialBackoff) ShouldRetry(err error, attempt int) bool {
	if err == nil {
		return false
	}

	if attempt >= e.Config.MaxAttempts {
		return false
	}

	// Check if error is categorized as transient
	if errors.IsTransient(err) {
		return true
	}

	// Check specific retryable errors if configured
	if len(e.Config.RetryableErrors) > 0 {
		return slices.Contains(e.Config.RetryableErrors, err)
	}

	// Don't retry permanent or business errors
	if errors.IsPermanent(err) || errors.IsBusiness(err) {
		return false
	}

	return false
}

// NextDelay calculates the next delay with exponential backoff and jitter
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(e.Config.InitialDelay) * math.Pow(e.Config.Multiplier, float64(attempt))

	// Cap at max delay
	maxDelay := float64(e.Config.MaxDelay)
	delay = math.Min(delay, maxDelay)

	// Add jitter: randomize +/- jitterFraction of the delay
	if e.Config.JitterFraction > 0 {
		jitterAmount := delay * e.Config.JitterFraction
		jitter := (rand.Float64() * 2 * jitterAmount) - jitterAmount
		delay += jitter
	}

	return time.Duration(delay)
}

// Do executes an operation with retry logic
func Do(ctx context.Context, strategy RetryStrategy, operation func() error) error {
	var lastErr error

	for attempt := 1; ; attempt++ {
		lastErr = operation()

		if lastErr == nil {
			return nil
		}

		if !strategy.ShouldRetry(lastErr, attempt) {
			return lastErr
		}

		delay := strategy.NextDelay(attempt)

		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return errors.NewTransientErrorf("retry cancelled: %w", ctx.Err())
		}
	}
}

// DoWithResult executes an operation that returns a value with retry logic
func DoWithResult[T any](ctx context.Context, strategy RetryStrategy, operation func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 1; ; attempt++ {
		result, lastErr = operation()

		if lastErr == nil {
			return result, nil
		}

		if !strategy.ShouldRetry(lastErr, attempt) {
			return result, lastErr
		}

		delay := strategy.NextDelay(attempt)

		t := time.NewTimer(delay)
		select {
		case <-t.C:
			// Continue to next attempt
		case <-ctx.Done():
			t.Stop()
			return result, errors.NewTransientErrorf("retry cancelled: %w", ctx.Err())
		}
	}
}
