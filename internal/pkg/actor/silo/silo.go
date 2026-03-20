package silo

import (
	"context"
	"sync"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const siloServiceName = "GrainSilo"

type SiloOptions struct {
	Timeout        time.Duration
	Telemetry      ports.TelemetryProvider
	RetrierFactory ports.RetrierFactoryInterface
}

func (opts *SiloOptions) Validate() error {
	if opts.Timeout <= 0 {
		return errors.NewValidationErrorf("timeout must be greater than zero")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider is required")
	}
	if opts.RetrierFactory == nil {
		return errors.NewValidationErrorf("retrier factory is required")
	}
	return nil
}

// Silo is the in-memory silo/cluster that hosts and manages grains
// Similar to Orleans silo - handles activation, deactivation, message routing
type Silo struct {
	activations sync.Map // map[string]*GrainActivation
	factories   sync.Map // map[string]ports.GrainFactory
	timeout     time.Duration
	retrier     ports.RetrierInterface
	telemetry   ports.TelemetryProvider
}

func NewSilo(opts *SiloOptions) (*Silo, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	strategy := opts.RetrierFactory.NewExponentialBackoffRetrier(siloServiceName, opts.Timeout)
	return &Silo{
		timeout:   opts.Timeout,
		telemetry: opts.Telemetry,
		retrier:   strategy,
	}, nil
}

// RegisterFactory registers a grain factory for a given kind
func (s *Silo) RegisterFactory(kind enums.AggregateType, factory ports.GrainFactory) {
	s.factories.Store(kind, factory)
}

// GetOrActivate retrieves or activates a grain
func (s *Silo) GetOrActivate(ctx context.Context, identity *grain.GrainIdentity) (ports.GrainRef, error) {
	ctx, span := s.TraceSpan(ctx, "GetOrActivate", identity)
	defer span.End()

	key := identity.String()
	if val, ok := s.activations.Load(key); ok {
		return val.(*GrainActivation), nil
	}

	factoryVal, ok := s.factories.Load(identity.Kind)
	if !ok {
		return nil, errors.NewPermanentErrorf("no factory registered for grain kind: %s", identity.Kind)
	}
	factory := factoryVal.(ports.GrainFactory)

	instance := factory(identity)

	activation := &GrainActivation{
		identity: identity,
		instance: instance,
	}

	actual, loaded := s.activations.LoadOrStore(key, activation)
	if loaded {
		return actual.(*GrainActivation), nil
	}

	err := s.retrier.Do(ctx, func(ctx context.Context) error {
		return instance.OnActivate(ctx, identity)
	})

	if err != nil {
		s.telemetry.Logger().ErrorContext(ctx, "grain activation failed after retry strategy",
			"identity", key,
			"error", err)
		s.activations.Delete(key)
		return nil, err
	}

	return activation, nil
}

// Tell sends a one-way message to a grain (fire and forget)
func (s *Silo) Tell(ctx context.Context, identity *grain.GrainIdentity, msg ports.Message) error {
	ctx, span := s.TraceSpan(ctx, "Tell", identity)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	activation, err := s.GetOrActivate(ctx, identity)
	if err != nil {
		return err
	}

	return activation.Tell(ctx, msg)
}

// Ask sends a request message and waits for a response
func (s *Silo) Ask(ctx context.Context, identity *grain.GrainIdentity, msg ports.Message) (ports.Message, error) {
	ctx, span := s.TraceSpan(ctx, "Ask", identity)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	activation, err := s.GetOrActivate(ctx, identity)
	if err != nil {
		return nil, err
	}

	return activation.Ask(ctx, msg)
}

// Deactivate removes a grain from memory
func (s *Silo) Deactivate(ctx context.Context, identity *grain.GrainIdentity) error {
	ctx, span := s.TraceSpan(ctx, "Deactivate", identity)
	defer span.End()

	key := identity.String()

	val, ok := s.activations.Load(key)
	if !ok {
		return nil 
	}

	activation := val.(*GrainActivation)
	err := activation.deactivate(ctx)
	if err != nil {
		s.telemetry.Logger().ErrorContext(ctx, "grain deactivation failed",
			"identity", key,
			"error", err)
		return err
	}

	s.activations.Delete(key)
	return nil
}

// Reset clears all activations
func (s *Silo) Reset(ctx context.Context) error {
	ctx, span := s.TraceSpan(ctx, "Reset", &grain.GrainIdentity{})
	defer span.End()

	var firstErr error
	s.activations.Range(func(key, value any) bool {
		activation := value.(*GrainActivation)
		if err := activation.deactivate(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})

	s.activations = sync.Map{}
	return firstErr
}

func (s *Silo) TraceSpan(ctx context.Context, operation string, identity *grain.GrainIdentity) (context.Context, trace.Span) {
	tracer := s.telemetry.Tracer()
	ctx, span := tracer.Start(ctx, siloServiceName+" "+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("grain.identity", identity.EntityID.String()),
			attribute.String("grain.kind", identity.Kind.String()),
		),
	)
	return ctx, span
}

var _ ports.Silo = (*Silo)(nil)
