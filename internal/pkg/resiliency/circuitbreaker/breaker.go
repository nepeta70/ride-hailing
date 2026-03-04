package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"

	pkgErrors "github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed   State = iota // Normal operation, requests pass through
	StateOpen                  // Circuit is open, requests fail fast
	StateHalfOpen              // Testing if service recovered
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

var (
	ErrCircuitOpen     = pkgErrors.NewTransientError("circuit breaker is open")
	ErrTooManyRequests = pkgErrors.NewTransientError("too many requests in half-open state")
	ErrPanicRecovered  = pkgErrors.NewTransientError("panic recovered in circuit breaker")
)

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures
// States: Closed -> Open -> Half-Open -> Closed
type CircuitBreaker struct {
	config          *CircuitBreakerConfig
	mu              sync.RWMutex
	state           State
	failures        uint32
	successes       uint32
	lastFailureTime time.Time
	halfOpenReqs    uint32 // Concurrent requests in half-open state
}

// NewCircuitBreaker creates a new CircuitBreaker with the given configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) (*CircuitBreaker, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}, nil
}

// Execute wraps the given function with circuit breaker logic
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) (err error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}

	isHalfOpen, beforeErr := cb.beforeRequest()
	if beforeErr != nil {
		return beforeErr
	}

	defer func() {
		if r := recover(); r != nil {
			// Convert panic to your existing Transient error type
			err = ErrPanicRecovered

			// Record the failure so the circuit actually transitions states
			cb.afterRequest(err, isHalfOpen)
		}
	}()

	// Execute the function and track result
	err = fn()
	cb.afterRequest(err, isHalfOpen)
	return err
}

// beforeRequest checks if request should proceed based on circuit state
// Returns (isHalfOpen, error)
func (cb *CircuitBreaker) beforeRequest() (bool, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if old failures should expire in closed state (TTL feature)
	if cb.state == StateClosed && cb.failures > 0 && !cb.lastFailureTime.IsZero() {
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.failures = 0
		}
	}

	switch cb.state {
	case StateClosed:
		return false, nil

	case StateOpen:
		// Check if timeout has elapsed to transition to half-open
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.setState(StateHalfOpen)
			cb.halfOpenReqs++
			return true, nil
		}
		return false, ErrCircuitOpen

	case StateHalfOpen:
		// Limit concurrent requests in half-open state
		if cb.halfOpenReqs >= cb.config.MaxRequests {
			return false, ErrTooManyRequests
		}
		cb.halfOpenReqs++
		return true, nil

	default:
		return false, ErrCircuitOpen
	}
}

// afterRequest updates circuit state based on request result
func (cb *CircuitBreaker) afterRequest(err error, isHalfOpen bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if isHalfOpen {
		cb.halfOpenReqs--
	}

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return
		}
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.successes = 0
	cb.lastFailureTime = time.Now().UTC()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.setState(StateOpen)
		}

	case StateHalfOpen:
		// Any failure in half-open immediately reopens circuit
		cb.setState(StateOpen)
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessesToClose {
			cb.setState(StateClosed)
		}
	}
}

// setState transitions to a new state and resets counters
func (cb *CircuitBreaker) setState(newState State) {
	cb.state = newState

	switch newState {
	case StateClosed:
		cb.failures = 0
		cb.successes = 0

	case StateOpen:
		cb.successes = 0
		// Keep failures for metrics

	case StateHalfOpen:
		cb.successes = 0
		cb.halfOpenReqs = 0
	}
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// IsOpen returns true if the circuit is currently open (rejecting requests)
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// Stats returns current statistics
type Stats struct {
	State       State
	Failures    uint32
	Successes   uint32
	LastFailure time.Time
}

// Stats returns current circuit breaker statistics
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return Stats{
		State:       cb.state,
		Failures:    cb.failures,
		Successes:   cb.successes,
		LastFailure: cb.lastFailureTime,
	}
}
