package circuitbreaker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	pkgErrors "github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/circuitbreaker"
	. "github.com/nepeta70/ride-hailing/internal/pkg/resiliency/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

var someConfig = &CircuitBreakerConfig{
	MaxFailures:      3,
	Timeout:          10 * time.Second,
	MaxRequests:      2,
	SuccessesToClose: 1,
}

var tel = mocks.NewMockTelemetryProvider()

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *CircuitBreakerConfig
	}{
		{
			name:   "with valid config",
			config: someConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCircuitBreaker(tt.config, tel)
			assert.NoError(t, err)
			assert.NotNil(t, cb)
			assert.Equal(t, StateClosed, cb.State())
		})
	}
}

func TestCircuitBreaker_Execute_TableDriven(t *testing.T) {
	type args struct {
		config *CircuitBreakerConfig
		fn     func() error
		ctx    context.Context
	}
	testErr := errors.New("test error")

	tests := []struct {
		name      string
		args      args
		before    func(cb *CircuitBreaker) // To setup state or simulate time
		wantErr   error
		wantState State
	}{
		{
			name: "success",
			args: args{
				config: &CircuitBreakerConfig{MaxFailures: 3, Timeout: 100 * time.Millisecond, MaxRequests: 1, SuccessesToClose: 2},
				fn:     func() error { return nil },
				ctx:    context.Background(),
			},
			wantErr:   nil,
			wantState: StateClosed,
		},
		{
			name: "failure",
			args: args{
				config: &CircuitBreakerConfig{MaxFailures: 3, Timeout: 100 * time.Millisecond, MaxRequests: 1, SuccessesToClose: 2},
				fn:     func() error { return testErr },
				ctx:    context.Background(),
			},
			wantErr:   testErr,
			wantState: StateClosed,
		},
		{
			name: "context canceled",
			args: args{
				config: DefaultConfig(),
				fn:     func() error { return nil },
				ctx:    func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			},
			wantErr:   context.Canceled,
			wantState: StateClosed,
		},
		{
			name: "panic recovery and remains closed",
			args: args{
				config: &CircuitBreakerConfig{MaxFailures: 2, Timeout: 100 * time.Millisecond, MaxRequests: 1, SuccessesToClose: 2},
				fn:     func() error { panic("test panic") },
				ctx:    context.Background(),
			},
			wantErr:   circuitbreaker.ErrPanicRecovered,
			wantState: StateClosed,
		},
		{
			name: "panic recovery triggers open",
			args: args{
				config: &CircuitBreakerConfig{MaxFailures: 1, Timeout: 100 * time.Millisecond, MaxRequests: 1, SuccessesToClose: 2},
				fn:     func() error { panic("test panic 2") },
				ctx:    context.Background(),
			},
			wantErr:   circuitbreaker.ErrPanicRecovered,
			wantState: StateOpen,
		},
		// --- TTL TEST CASES ---
		{
			name: "failure TTL expires: count resets",
			args: args{
				config: &CircuitBreakerConfig{MaxFailures: 2, Timeout: 50 * time.Millisecond, MaxRequests: 1, SuccessesToClose: 2},
				fn:     func() error { return testErr },
				ctx:    context.Background(),
			},
			before: func(cb *CircuitBreaker) {
				// Record first failure
				_ = cb.Execute(context.Background(), func() error { return testErr })
				// Wait for the TTL window to expire
				time.Sleep(60 * time.Millisecond)
			},
			wantErr:   testErr,
			wantState: StateClosed, // Should stay closed because first failure was "forgotten"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCircuitBreaker(tt.args.config, tel)
			assert.NoError(t, err)
			assert.NotNil(t, cb)

			if tt.before != nil {
				tt.before(cb)
			}

			err = cb.Execute(tt.args.ctx, tt.args.fn)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}

			assert.Equal(t, tt.wantState, cb.State())
		})
	}
}

func TestCircuitBreaker_StateTransition_ClosedToOpen(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      3,
		Timeout:          100 * time.Millisecond,
		MaxRequests:      1,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	testErr := errors.New("test error")

	// Circuit should start closed
	assert.Equal(t, StateClosed, cb.State())

	// First two failures should keep circuit closed
	for range 2 {
		err := cb.Execute(context.Background(), func() error {
			return testErr
		})
		assert.Error(t, err)
		assert.Equal(t, StateClosed, cb.State())
	}

	// Third failure should open circuit
	err = cb.Execute(context.Background(), func() error {
		return testErr
	})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_StateTransition_OpenToHalfOpen(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		MaxRequests:      1,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	testErr := errors.New("test error")

	// Open the circuit
	for range 2 {
		cb.Execute(context.Background(), func() error {
			return testErr
		})
	}
	assert.Equal(t, StateOpen, cb.State())

	// Immediate request should fail
	err = cb.Execute(context.Background(), func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitOpen, err)

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next request should transition to half-open
	err = cb.Execute(context.Background(), func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.State())
}

func TestCircuitBreaker_StateTransition_HalfOpenToClosed(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		MaxRequests:      2,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	testErr := errors.New("test error")

	// Open the circuit
	for range 2 {
		cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	// Wait and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.Execute(context.Background(), func() error {
		return nil
	})
	assert.Equal(t, StateHalfOpen, cb.State())

	// Second success should close circuit
	err = cb.Execute(context.Background(), func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_StateTransition_HalfOpenToOpen(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		MaxRequests:      2,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	// Wait and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.Execute(context.Background(), func() error {
		return nil
	})
	assert.Equal(t, StateHalfOpen, cb.State())

	// Any failure in half-open should reopen circuit
	err = cb.Execute(context.Background(), func() error {
		return testErr
	})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_HalfOpen_MaxRequests(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		MaxRequests:      1,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	testErr := errors.New("test error")

	// Open the circuit
	for range 2 {
		cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	// Wait and transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Use a slow request to keep half-open state occupied
	var wg sync.WaitGroup
	started := make(chan struct{})
	wg.Go(func() {
		cb.Execute(context.Background(), func() error {
			close(started) // Signal that we're executing
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	})

	// Wait for the first request to actually start executing
	<-started

	// Second concurrent request should fail with ErrTooManyRequests
	err = cb.Execute(context.Background(), func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrTooManyRequests, err)

	wg.Wait()
}

func TestCircuitBreaker_ContextErrors_NotCountedAsFailures(t *testing.T) {
	tests := []struct {
		name        string
		contextErr  error
		setupCtx    func() context.Context
		description string
	}{
		{
			name:       "deadline exceeded",
			contextErr: context.DeadlineExceeded,
			setupCtx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()
				time.Sleep(5 * time.Millisecond)
				return ctx
			},
			description: "deadline exceeded should not count as failure",
		},
		{
			name:       "canceled",
			contextErr: context.Canceled,
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			description: "canceled context should not count as failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
				MaxFailures:      2,
				Timeout:          100 * time.Millisecond,
				MaxRequests:      1,
				SuccessesToClose: 2,
			}, tel)
			assert.NoError(t, err)

			// Execute with context error
			err = cb.Execute(context.Background(), func() error {
				return tt.contextErr
			})

			// Circuit should remain closed
			assert.Equal(t, StateClosed, cb.State())

			// Stats should show no failures
			stats := cb.Stats()
			assert.Equal(t, uint32(0), stats.Failures)
		})
	}
}

func TestCircuitBreaker_IsOpen(t *testing.T) {
	tests := []struct {
		name     string
		state    State
		expected bool
	}{
		{
			name:     "closed circuit",
			state:    StateClosed,
			expected: false,
		},
		{
			name:     "open circuit",
			state:    StateOpen,
			expected: true,
		},
		{
			name:     "half-open circuit",
			state:    StateHalfOpen,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
				MaxFailures:      2,
				Timeout:          50 * time.Millisecond,
				MaxRequests:      1,
				SuccessesToClose: 2,
			}, tel)
			assert.NoError(t, err)

			// Force state
			switch tt.state {
			case StateOpen:
				testErr := errors.New("test error")
				for range 2 {
					cb.Execute(context.Background(), func() error {
						return testErr
					})
				}
			case StateHalfOpen:
				testErr := errors.New("test error")
				for range 2 {
					cb.Execute(context.Background(), func() error {
						return testErr
					})
				}
				time.Sleep(60 * time.Millisecond)
				cb.Execute(context.Background(), func() error {
					return nil
				})
			}

			assert.Equal(t, tt.expected, cb.IsOpen())
		})
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      3,
		Timeout:          100 * time.Millisecond,
		MaxRequests:      1,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	// Initial stats
	stats := cb.Stats()
	assert.Equal(t, StateClosed, stats.State)
	assert.Equal(t, uint32(0), stats.Failures)
	assert.Equal(t, uint32(0), stats.Successes)

	// After one failure
	testErr := errors.New("test error")
	cb.Execute(context.Background(), func() error {
		return testErr
	})

	stats = cb.Stats()
	assert.Equal(t, StateClosed, stats.State)
	assert.Equal(t, uint32(1), stats.Failures)
	assert.Equal(t, uint32(0), stats.Successes)
	assert.False(t, stats.LastFailure.IsZero())

	// After success
	cb.Execute(context.Background(), func() error {
		return nil
	})

	stats = cb.Stats()
	assert.Equal(t, uint32(0), stats.Failures) // Failures reset on success in closed state
}

func TestState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    State
		expected string
	}{
		{
			name:     "closed state",
			state:    StateClosed,
			expected: "closed",
		},
		{
			name:     "open state",
			state:    StateOpen,
			expected: "open",
		},
		{
			name:     "half-open state",
			state:    StateHalfOpen,
			expected: "half-open",
		},
		{
			name:     "unknown state",
			state:    State(99),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      10,
		Timeout:          100 * time.Millisecond,
		MaxRequests:      5,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	concurrency := 50

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var err error
			if id%3 == 0 {
				err = errors.New("test error")
			}
			cb.Execute(context.Background(), func() error {
				return err
			})
		}(i)
	}

	wg.Wait()

	// Just verify no race conditions occurred
	stats := cb.Stats()
	assert.NotNil(t, stats)
}

func TestCircuitBreaker_SuccessResetsFailuresInClosedState(t *testing.T) {
	cb, err := NewCircuitBreaker(&CircuitBreakerConfig{
		MaxFailures:      3,
		Timeout:          100 * time.Millisecond,
		MaxRequests:      1,
		SuccessesToClose: 2,
	}, tel)
	assert.NoError(t, err)

	testErr := errors.New("test error")

	// Two failures
	for range 2 {
		cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	stats := cb.Stats()
	assert.Equal(t, uint32(2), stats.Failures)

	// One success should reset failures
	cb.Execute(context.Background(), func() error {
		return nil
	})

	stats = cb.Stats()
	assert.Equal(t, uint32(0), stats.Failures)
	assert.Equal(t, StateClosed, stats.State)
}

func TestCircuitBreaker_Errors(t *testing.T) {
	t.Run("ErrCircuitOpen is transient", func(t *testing.T) {
		assert.True(t, pkgErrors.IsTransient(ErrCircuitOpen))
	})

	t.Run("ErrTooManyRequests is transient", func(t *testing.T) {
		assert.True(t, pkgErrors.IsTransient(ErrTooManyRequests))
	})
}
