package circuitbreaker

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures      uint32        // Number of failures before opening
	Timeout          time.Duration // Time to wait before transitioning from Open to Half-Open
	MaxRequests      uint32        // Max concurrent requests allowed in Half-Open state
	SuccessesToClose uint32        // Successes needed in Half-Open to close circuit
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:      5,
		Timeout:          30 * time.Second,
		MaxRequests:      1,
		SuccessesToClose: 2,
	}
}

func (c *CircuitBreakerConfig) Validate() error {
	if c.MaxFailures == 0 {
		return errors.NewValidationErrorf("MaxFailures must be greater than 0")
	}
	if c.Timeout <= 0 {
		return errors.NewValidationErrorf("Timeout must be greater than 0")
	}
	if c.MaxRequests == 0 {
		return errors.NewValidationErrorf("MaxRequests must be greater than 0")
	}
	if c.SuccessesToClose == 0 {
		return errors.NewValidationErrorf("SuccessesToClose must be greater than 0")
	}

	return nil
}
