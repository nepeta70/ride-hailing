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
	config   RetryConfig
	logger   ports.Logger
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
		result, lastErr = op()
		if lastErr == nil {
			if attempt > 1 {
				r.logger.Info("operation succeeded after %d attempts", attempt)
			}
			return result, nil
		}

		if !r.shouldRetry(lastErr, attempt) {
			if attempt > 1 {
				r.logger.Warn("operation failed after %d attempts: %v", attempt, lastErr)
			}
			return result, lastErr
		}

		delay := r.strategy.NextDelay(attempt)
		r.logger.Warn("operation failed on attempt %d: %v. Retrying in %s...", attempt, lastErr, delay)

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return result, errors.NewTransientErrorf("retry cancelled: %w", ctx.Err())
		}
	}
}
