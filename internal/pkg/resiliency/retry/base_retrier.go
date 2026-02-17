package retry

import (
	"context"
	"slices"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type baseRetrier struct {
	strategy RetryStrategy
	config   *RetryConfig
	logger   ports.Logger
	observer ports.RetryObserver
}

// ShouldRetry determines if an error is retryable
func (r *baseRetrier) shouldRetry(err error, attempt int) bool {
	if err == nil {
		return false
	}

	if attempt >= r.config.MaxAttempts {
		return false
	}

	// Check if error is categorized as transient
	if errors.IsTransient(err) {
		return true
	}

	// Check specific retryable errors if configured
	if len(r.config.RetryableErrors) > 0 {
		return slices.Contains(r.config.RetryableErrors, err)
	}

	// Don't retry permanent or business errors
	if errors.IsPermanent(err) || errors.IsBusiness(err) {
		return false
	}

	return false
}

func (r *baseRetrier) DoWithResult(ctx context.Context, op func() (any, error)) (any, error) {
	var result any
	var lastErr error

	for attempt := 1; ; attempt++ {
		r.logger.Debug("Attempt for operation", "attempt", attempt)
		result, lastErr = op()
		if lastErr == nil {
			if attempt > 1 {
				r.logger.Info("operation succeeded after", "attempts", attempt)
			}
			return result, nil
		}

		if !r.shouldRetry(lastErr, attempt) {
			if attempt > 1 {
				r.logger.Warn("operation failed after", "attempts", attempt, "error", lastErr)
			}
			return result, lastErr
		}

		delay := r.strategy.NextDelay(attempt)

		if r.observer != nil {
			r.observer.ObserveRetry(attempt, lastErr, delay)
		}

		r.logger.Warn("operation failed on attempt", "attempt", attempt, "error", lastErr, "retry_in", delay)
		if r.observer != nil {
			r.observer.ObserveRetry(attempt, lastErr, delay)
		}

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return result, errors.NewTransientErrorf("retry cancelled: %w", ctx.Err())
		}
	}
}
