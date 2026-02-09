package retry

import "time"

// RetryStrategy defines how to handle retries
type RetryStrategy interface {
	NextDelay(attempt int) time.Duration
}
