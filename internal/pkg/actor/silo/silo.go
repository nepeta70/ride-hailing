package silo

import (
	"context"
	"sync"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/retry"
)

// GrainActivation holds an active grain instance
type GrainActivation struct {
	Identity *grain.GrainIdentity
	Instance ports.Grain
	mu       sync.RWMutex
}

type SiloOptions struct {
	Timeout time.Duration
	Logger  ports.Logger
}

// Silo is the in-memory silo/cluster that hosts and manages grains
// Similar to Orleans silo - handles activation, deactivation, message routing
type Silo struct {
	activations sync.Map // map[string]*GrainActivation
	factories   sync.Map // map[string]ports.GrainFactory
	timeout     time.Duration
	retrier     *retry.Retrier
	logger      ports.Logger
}

func NewSilo(opts *SiloOptions) *Silo {
	strategy := retry.NewExponentialBackoffRetrierWithTimeout(opts.Timeout, opts.Logger)
	return &Silo{
		timeout: opts.Timeout,
		logger:  opts.Logger,
		retrier: strategy,
	}
}

// RegisterFactory registers a grain factory for a given kind
func (s *Silo) RegisterFactory(kind enums.AggregateType, factory ports.GrainFactory) {
	s.factories.Store(kind, factory)
}

// GetOrActivate retrieves or activates a grain
func (s *Silo) GetOrActivate(ctx context.Context, identity *grain.GrainIdentity) (*GrainActivation, error) {
	key := identity.String()

	// Fast path: already activated
	if val, ok := s.activations.Load(key); ok {
		return val.(*GrainActivation), nil
	}

	// Get factory
	factoryVal, ok := s.factories.Load(identity.Kind)
	if !ok {
		return nil, errors.NewPermanentErrorf("no factory registered for grain kind: %s", identity.Kind)
	}
	factory := factoryVal.(ports.GrainFactory)

	// Create instance with identity injected
	instance := factory(identity)

	activation := &GrainActivation{
		Identity: identity,
		Instance: instance,
	}

	// Try to store (might lose race)
	actual, loaded := s.activations.LoadOrStore(key, activation)
	if loaded {
		// Lost race, use existing
		return actual.(*GrainActivation), nil
	}

	err := s.retrier.Do(ctx, func() error {
		return instance.OnActivate(ctx, identity)
	})

	// Won race, activate (pass identity so grain knows which entity it is)
	if err != nil {
		s.logger.Error("grain activation failed after retry strategy",
			"identity", key,
			"error", err)
		s.activations.Delete(key) // Cleanup so next request can try fresh
		return nil, err
	}

	return activation, nil
}

// Tell sends a one-way message to a grain (fire and forget)
func (s *Silo) Tell(ctx context.Context, identity *grain.GrainIdentity, msg ports.Message) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	activation, err := s.GetOrActivate(ctx, identity)
	if err != nil {
		return err
	}

	activation.mu.Lock()
	defer activation.mu.Unlock()

	_, err = activation.Instance.OnReceive(ctx, msg)
	return err
}

// Ask sends a request message and waits for a response
func (s *Silo) Ask(ctx context.Context, identity *grain.GrainIdentity, msg ports.Message) (ports.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	activation, err := s.GetOrActivate(ctx, identity)
	if err != nil {
		return nil, err
	}

	activation.mu.Lock()
	defer activation.mu.Unlock()

	return activation.Instance.OnReceive(ctx, msg)
}

// Deactivate removes a grain from memory
func (s *Silo) Deactivate(ctx context.Context, identity *grain.GrainIdentity) error {
	key := identity.String()

	val, ok := s.activations.Load(key)
	if !ok {
		return nil // already deactivated
	}

	activation := val.(*GrainActivation)
	activation.mu.Lock()
	defer activation.mu.Unlock()

	if err := activation.Instance.OnDeactivate(ctx); err != nil {
		return err
	}

	s.activations.Delete(key)
	return nil
}

// Reset clears all activations
func (s *Silo) Reset(ctx context.Context) error {
	var firstErr error
	s.activations.Range(func(key, value any) bool {
		activation := value.(*GrainActivation)
		if err := activation.Instance.OnDeactivate(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})

	s.activations = sync.Map{}
	return firstErr
}
