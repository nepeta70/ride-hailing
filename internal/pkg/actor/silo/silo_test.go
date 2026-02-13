package silo_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	. "github.com/nepeta70/ride-hailing/internal/pkg/actor/silo"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/mocks"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

const (
	driverKind enums.AggregateType = "driver"
	userKind   enums.AggregateType = "user"
	rideKind   enums.AggregateType = "ride"
)

var kinds = []enums.AggregateType{rideKind, driverKind, userKind}

// Mock Grain
type mockGrain struct {
	identity         *grain.GrainIdentity
	activateCalled   bool
	deactivateCalled bool
	receiveCalled    bool
	activateErr      error
	deactivateErr    error
	receiveErr       error
	receiveResponse  ports.Message
	onReceiveFunc    func(ctx context.Context, msg ports.Message) (ports.Message, error)
	mu               sync.Mutex
}

func (m *mockGrain) OnActivate(ctx context.Context, identity *grain.GrainIdentity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activateCalled = true
	m.identity = identity
	return m.activateErr
}

func (m *mockGrain) OnReceive(ctx context.Context, msg ports.Message) (ports.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receiveCalled = true
	if m.onReceiveFunc != nil {
		return m.onReceiveFunc(ctx, msg)
	}
	return m.receiveResponse, m.receiveErr
}

func (m *mockGrain) OnDeactivate(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deactivateCalled = true
	return m.deactivateErr
}

func (m *mockGrain) getActivateCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activateCalled
}

func (m *mockGrain) getDeactivateCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deactivateCalled
}

func (m *mockGrain) getReceiveCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.receiveCalled
}

// Test message types
type testMessage struct {
	data string
}

type testResponse struct {
	result string
}

func TestNewSilo(t *testing.T) {
	tests := []struct {
		name    string
		opts    *SiloOptions
		wantNil bool
	}{
		{
			name: "creates silo with valid options",
			opts: &SiloOptions{
				Timeout: 5 * time.Second,
				Logger:  &mocks.MockLogger{},
			},
			wantNil: false,
		},
		{
			name: "creates silo with minimal timeout",
			opts: &SiloOptions{
				Timeout: 1 * time.Millisecond,
				Logger:  &mocks.MockLogger{},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silo, _ := NewSilo(tt.opts)
			if tt.wantNil {
				assert.Nil(t, silo)
			} else {
				assert.NotNil(t, silo)
			}
		})
	}
}

func TestSilo_RegisterFactory(t *testing.T) {
	tests := []struct {
		name    string
		kind    enums.AggregateType
		factory ports.GrainFactory
	}{
		{
			name: "registers factory for ride aggregate",
			kind: rideKind,
			factory: func(identity *grain.GrainIdentity) ports.Grain {
				return &mockGrain{}
			},
		},
		{
			name: "registers factory for driver aggregate",
			kind: driverKind,
			factory: func(identity *grain.GrainIdentity) ports.Grain {
				return &mockGrain{}
			},
		},
		{
			name: "registers factory for user aggregate",
			kind: userKind,
			factory: func(identity *grain.GrainIdentity) ports.Grain {
				return &mockGrain{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silo, _ := NewSilo(&SiloOptions{
				Timeout: 5 * time.Second,
				Logger:  &mocks.MockLogger{},
			})

			silo.RegisterFactory(tt.kind, tt.factory)

			// Verify factory is registered by trying to activate a grain
			identity := grain.NewGrainIdentity(tt.kind, uuid.New())
			activation, err := silo.GetOrActivate(context.Background(), identity)

			assert.NoError(t, err)
			assert.NotNil(t, activation)
			assert.Equal(t, identity, activation.Identity)
		})
	}
}

func TestSilo_GetOrActivate(t *testing.T) {
	tests := []struct {
		name          string
		kind          enums.AggregateType
		setupFactory  bool
		activateErr   error
		wantErr       bool
		errContains   string
		validateGrain func(t *testing.T, g *mockGrain)
	}{
		{
			name:         "successfully activates new grain",
			kind:         rideKind,
			setupFactory: true,
			activateErr:  nil,
			wantErr:      false,
			validateGrain: func(t *testing.T, g *mockGrain) {
				assert.True(t, g.getActivateCalled())
			},
		},
		{
			name:         "returns existing grain on second call",
			kind:         rideKind,
			setupFactory: true,
			activateErr:  nil,
			wantErr:      false,
		},
		{
			name:         "returns error when factory not registered",
			kind:         rideKind,
			setupFactory: false,
			wantErr:      true,
			errContains:  "no factory registered",
		},
		{
			name:         "returns error when activation fails",
			kind:         rideKind,
			setupFactory: true,
			activateErr:  errors.New("activation failed"),
			wantErr:      true,
			errContains:  "activation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silo, _ := NewSilo(&SiloOptions{
				Timeout: 5 * time.Second,
				Logger:  &mocks.MockLogger{},
			})

			var testGrain *mockGrain
			if tt.setupFactory {
				testGrain = &mockGrain{activateErr: tt.activateErr}
				factory := func(identity *grain.GrainIdentity) ports.Grain {
					return testGrain
				}
				silo.RegisterFactory(tt.kind, factory)
			}

			identity := grain.NewGrainIdentity(tt.kind, uuid.New())
			activation, err := silo.GetOrActivate(context.Background(), identity)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				if tt.activateErr == nil {
					assert.Nil(t, activation)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, activation)
				assert.Equal(t, identity, activation.Identity)

				if tt.validateGrain != nil && testGrain != nil {
					tt.validateGrain(t, testGrain)
				}

				// Test idempotency - second call should return same activation
				activation2, err2 := silo.GetOrActivate(context.Background(), identity)
				assert.NoError(t, err2)
				assert.Equal(t, activation, activation2)
			}
		})
	}
}

func TestSilo_Tell(t *testing.T) {
	tests := []struct {
		name         string
		setupFactory bool
		activateErr  error
		receiveErr   error
		wantErr      bool
		errContains  string
		validateCall func(t *testing.T, g *mockGrain)
	}{
		{
			name:         "successfully sends message",
			setupFactory: true,
			activateErr:  nil,
			receiveErr:   nil,
			wantErr:      false,
			validateCall: func(t *testing.T, g *mockGrain) {
				assert.True(t, g.getReceiveCalled())
			},
		},
		{
			name:         "returns error when grain activation fails",
			setupFactory: true,
			activateErr:  errors.New("activation failed"),
			wantErr:      true,
			errContains:  "activation failed",
		},
		{
			name:         "returns error when factory not registered",
			setupFactory: false,
			wantErr:      true,
			errContains:  "no factory registered",
		},
		{
			name:         "returns error when receive fails",
			setupFactory: true,
			receiveErr:   errors.New("receive failed"),
			wantErr:      true,
			errContains:  "receive failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silo, _ := NewSilo(&SiloOptions{
				Timeout: 5 * time.Second,
				Logger:  &mocks.MockLogger{},
			})

			var testGrain *mockGrain
			if tt.setupFactory {
				testGrain = &mockGrain{
					activateErr: tt.activateErr,
					receiveErr:  tt.receiveErr,
				}
				factory := func(identity *grain.GrainIdentity) ports.Grain {
					return testGrain
				}
				silo.RegisterFactory(rideKind, factory)
			}

			identity := grain.NewGrainIdentity(rideKind, uuid.New())
			msg := testMessage{data: "test message"}

			err := silo.Tell(context.Background(), identity, msg)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateCall != nil && testGrain != nil {
					tt.validateCall(t, testGrain)
				}
			}
		})
	}
}

func TestSilo_Ask(t *testing.T) {
	tests := []struct {
		name            string
		setupFactory    bool
		activateErr     error
		receiveErr      error
		receiveResponse ports.Message
		wantErr         bool
		errContains     string
		validateCall    func(t *testing.T, g *mockGrain, resp ports.Message)
	}{
		{
			name:            "successfully sends request and receives response",
			setupFactory:    true,
			activateErr:     nil,
			receiveErr:      nil,
			receiveResponse: testResponse{result: "success"},
			wantErr:         false,
			validateCall: func(t *testing.T, g *mockGrain, resp ports.Message) {
				assert.True(t, g.getReceiveCalled())
				assert.NotNil(t, resp)
				response, ok := resp.(testResponse)
				assert.True(t, ok)
				assert.Equal(t, "success", response.result)
			},
		},
		{
			name:         "returns error when grain activation fails",
			setupFactory: true,
			activateErr:  errors.New("activation failed"),
			wantErr:      true,
			errContains:  "activation failed",
		},
		{
			name:         "returns error when factory not registered",
			setupFactory: false,
			wantErr:      true,
			errContains:  "no factory registered",
		},
		{
			name:         "returns error when receive fails",
			setupFactory: true,
			receiveErr:   errors.New("receive failed"),
			wantErr:      true,
			errContains:  "receive failed",
		},
		{
			name:            "handles nil response",
			setupFactory:    true,
			receiveResponse: nil,
			wantErr:         false,
			validateCall: func(t *testing.T, g *mockGrain, resp ports.Message) {
				assert.True(t, g.getReceiveCalled())
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silo, _ := NewSilo(&SiloOptions{
				Timeout: 5 * time.Second,
				Logger:  &mocks.MockLogger{},
			})

			var testGrain *mockGrain
			if tt.setupFactory {
				testGrain = &mockGrain{
					activateErr:     tt.activateErr,
					receiveErr:      tt.receiveErr,
					receiveResponse: tt.receiveResponse,
				}
				factory := func(identity *grain.GrainIdentity) ports.Grain {
					return testGrain
				}
				silo.RegisterFactory(rideKind, factory)
			}

			identity := grain.NewGrainIdentity(rideKind, uuid.New())
			msg := testMessage{data: "test request"}

			resp, err := silo.Ask(context.Background(), identity, msg)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateCall != nil && testGrain != nil {
					tt.validateCall(t, testGrain, resp)
				}
			}
		})
	}
}

func TestSilo_Deactivate(t *testing.T) {
	tests := []struct {
		name           string
		preActivate    bool
		deactivateErr  error
		wantErr        bool
		errContains    string
		validateCall   func(t *testing.T, g *mockGrain)
		testIdempotent bool
	}{
		{
			name:        "successfully deactivates active grain",
			preActivate: true,
			wantErr:     false,
			validateCall: func(t *testing.T, g *mockGrain) {
				assert.True(t, g.getDeactivateCalled())
			},
		},
		{
			name:           "deactivating non-existent grain is no-op",
			preActivate:    false,
			wantErr:        false,
			testIdempotent: true,
		},
		{
			name:          "returns error when deactivation fails",
			preActivate:   true,
			deactivateErr: errors.New("deactivation failed"),
			wantErr:       true,
			errContains:   "deactivation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silo, _ := NewSilo(&SiloOptions{
				Timeout: 5 * time.Second,
				Logger:  &mocks.MockLogger{},
			})

			var testGrain *mockGrain
			identity := grain.NewGrainIdentity(rideKind, uuid.New())

			if tt.preActivate {
				testGrain = &mockGrain{deactivateErr: tt.deactivateErr}
				factory := func(identity *grain.GrainIdentity) ports.Grain {
					return testGrain
				}
				silo.RegisterFactory(rideKind, factory)

				_, err := silo.GetOrActivate(context.Background(), identity)
				assert.NoError(t, err)
			}

			err := silo.Deactivate(context.Background(), identity)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateCall != nil && testGrain != nil {
					tt.validateCall(t, testGrain)
				}
			}

			// Test idempotency if requested
			if tt.testIdempotent {
				err2 := silo.Deactivate(context.Background(), identity)
				assert.NoError(t, err2)
			}
		})
	}
}

func TestSilo_Reset(t *testing.T) {
	tests := []struct {
		name               string
		grainCount         int
		deactivateErr      error
		deactivateErrIndex int
		wantErr            bool
		validateCall       func(t *testing.T, grains []*mockGrain)
	}{
		{
			name:       "successfully resets silo with no grains",
			grainCount: 0,
			wantErr:    false,
		},
		{
			name:       "successfully resets silo with single grain",
			grainCount: 1,
			wantErr:    false,
			validateCall: func(t *testing.T, grains []*mockGrain) {
				assert.True(t, grains[0].getDeactivateCalled())
			},
		},
		{
			name:       "successfully resets silo with multiple grains",
			grainCount: 3,
			wantErr:    false,
			validateCall: func(t *testing.T, grains []*mockGrain) {
				for _, g := range grains {
					assert.True(t, g.getDeactivateCalled())
				}
			},
		},
		{
			name:               "returns error when deactivation fails",
			grainCount:         3,
			deactivateErr:      errors.New("deactivation failed"),
			deactivateErrIndex: 1,
			wantErr:            true,
			validateCall: func(t *testing.T, grains []*mockGrain) {
				// All grains should still be attempted to deactivate
				for _, g := range grains {
					assert.True(t, g.getDeactivateCalled())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silo, _ := NewSilo(&SiloOptions{
				Timeout: 5 * time.Second,
				Logger:  &mocks.MockLogger{},
			})

			grains := make([]*mockGrain, tt.grainCount)

			// Activate grains

			for i := 0; i < tt.grainCount; i++ {
				idx := i
				grains[i] = &mockGrain{}
				if tt.deactivateErr != nil && i == tt.deactivateErrIndex {
					grains[i].deactivateErr = tt.deactivateErr
				}

				factory := func(identity *grain.GrainIdentity) ports.Grain {
					return grains[idx]
				}

				kind := kinds[i%len(kinds)]
				silo.RegisterFactory(kind, factory)

				identity := grain.NewGrainIdentity(kind, uuid.New())
				_, err := silo.GetOrActivate(context.Background(), identity)
				assert.NoError(t, err)
			}

			err := silo.Reset(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.validateCall != nil {
				tt.validateCall(t, grains)
			}
		})
	}
}

func TestSilo_Concurrency(t *testing.T) {
	t.Run("concurrent activations of same grain", func(t *testing.T) {
		silo, _ := NewSilo(&SiloOptions{
			Timeout: 5 * time.Second,
			Logger:  &mocks.MockLogger{},
		})

		activationCount := 0
		var mu sync.Mutex

		factory := func(identity *grain.GrainIdentity) ports.Grain {
			mu.Lock()
			activationCount++
			mu.Unlock()
			return &mockGrain{}
		}

		silo.RegisterFactory(rideKind, factory)

		identity := grain.NewGrainIdentity(rideKind, uuid.New())

		// Concurrent activations
		var wg sync.WaitGroup
		concurrency := 10
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				_, err := silo.GetOrActivate(context.Background(), identity)
				assert.NoError(t, err)
			}()
		}

		wg.Wait()

		// Only one activation should occur
		mu.Lock()
		assert.Equal(t, 1, activationCount)
		mu.Unlock()
	})

	t.Run("concurrent tell operations", func(t *testing.T) {
		silo, _ := NewSilo(&SiloOptions{
			Timeout: 5 * time.Second,
			Logger:  &mocks.MockLogger{},
		})

		receiveCount := 0
		var mu sync.Mutex

		testGrain := &mockGrain{
			onReceiveFunc: func(ctx context.Context, msg ports.Message) (ports.Message, error) {
				mu.Lock()
				receiveCount++
				mu.Unlock()
				return nil, nil
			},
		}

		factory := func(identity *grain.GrainIdentity) ports.Grain {
			return testGrain
		}

		silo.RegisterFactory(rideKind, factory)

		identity := grain.NewGrainIdentity(rideKind, uuid.New())

		// Concurrent tells
		var wg sync.WaitGroup
		concurrency := 10
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				err := silo.Tell(context.Background(), identity, testMessage{data: "test"})
				assert.NoError(t, err)
			}()
		}

		wg.Wait()

		// All messages should be received
		mu.Lock()
		assert.Equal(t, concurrency, receiveCount)
		mu.Unlock()
	})
}

func TestSilo_ContextTimeout(t *testing.T) {
	t.Run("tell respects context timeout", func(t *testing.T) {
		silo, _ := NewSilo(&SiloOptions{
			Timeout: 1 * time.Millisecond,
			Logger:  &mocks.MockLogger{},
		})

		testGrain := &mockGrain{
			onReceiveFunc: func(ctx context.Context, msg ports.Message) (ports.Message, error) {
				// Simulate slow processing that respects context
				select {
				case <-time.After(100 * time.Millisecond):
					return nil, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}

		factory := func(identity *grain.GrainIdentity) ports.Grain {
			return testGrain
		}

		silo.RegisterFactory(rideKind, factory)

		identity := grain.NewGrainIdentity(rideKind, uuid.New())

		err := silo.Tell(context.Background(), identity, testMessage{data: "test"})

		// Should timeout
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("ask respects context timeout", func(t *testing.T) {
		silo, _ := NewSilo(&SiloOptions{
			Timeout: 1 * time.Millisecond,
			Logger:  &mocks.MockLogger{},
		})

		testGrain := &mockGrain{
			onReceiveFunc: func(ctx context.Context, msg ports.Message) (ports.Message, error) {
				// Simulate slow processing that respects context
				select {
				case <-time.After(100 * time.Millisecond):
					return testResponse{result: "success"}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}

		factory := func(identity *grain.GrainIdentity) ports.Grain {
			return testGrain
		}

		silo.RegisterFactory(rideKind, factory)

		identity := grain.NewGrainIdentity(rideKind, uuid.New())

		_, err := silo.Ask(context.Background(), identity, testMessage{data: "test"})

		// Should timeout
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}
