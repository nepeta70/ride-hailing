package retry

import (
	"context"
	"slices"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/trace"
)

type baseRetrier struct {
	serviceName string
	strategy    RetryStrategy
	config      *RetryConfig
	telemetry   ports.TelemetryProvider
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

func (r *baseRetrier) DoWithResult(ctx context.Context, op func(ctx context.Context) (any, error)) (any, error) {
	ctx, span := r.telemetry.Tracer().Start(ctx, "Retry Loop: "+r.serviceName,
		trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	var result any
	var lastErr error

	for attempt := 1; ; attempt++ {
		r.telemetry.Logger().DebugContext(ctx, "Attempt for operation", "attempt", attempt)
		result, lastErr = op(ctx)
		if lastErr == nil {
			if attempt > 1 {
				r.telemetry.Logger().DebugContext(ctx, "operation succeeded after", "attempts", attempt)
			}
			return result, nil
		}

		if !r.shouldRetry(lastErr, attempt) {
			if attempt > 1 {
				r.telemetry.Logger().WarnContext(ctx, "operation failed after", "attempts", attempt, "error", lastErr)
			}
			return result, lastErr
		}

		delay := r.strategy.NextDelay(attempt)

		r.telemetry.Logger().WarnContext(ctx, "operation failed on attempt", "attempt", attempt, "error", lastErr, "retry_in", delay)
		r.telemetry.Metrics().ObserveRetry(r.serviceName, attempt, lastErr, delay)

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return result, errors.NewTransientErrorf("retry cancelled: %w", ctx.Err())
		}
	}
}
