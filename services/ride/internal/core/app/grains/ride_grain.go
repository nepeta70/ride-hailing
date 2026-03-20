package grains

import (
	"context"
	"reflect"
	"slices"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/shopspring/decimal"

	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RideGrainOptions struct {
	Storage        ports.GrainStorage
	EventPub       pkgPorts.EventPublisher
	Topic          contracts.Topic
	ContextManager *ctxmgr.ContextManager
	Telemetry      pkgPorts.TelemetryProvider
}

func (opts *RideGrainOptions) Validate() error {
	if opts.Storage == nil {
		return errors.NewValidationErrorf("grain storage is required")
	}
	if opts.EventPub == nil {
		return errors.NewValidationErrorf("event publisher is required")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider is required")
	}
	if opts.Topic == "" {
		return errors.NewValidationErrorf("topic is required")
	}
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	return nil
}

var terminalStates = []domain.RideStatus{
	domain.RideStatusCompleted,
	domain.RideStatusCancelled,
}

type RideGrain struct {
	identity       *grain.GrainIdentity
	core           *domain.RideCore
	state          *domain.RideState
	version        int
	storage        ports.GrainStorage
	eventPub       pkgPorts.EventPublisher
	telemetry      pkgPorts.TelemetryProvider
	topic          contracts.Topic
	contextManager *ctxmgr.ContextManager
}

var _ pkgPorts.Grain = (*RideGrain)(nil)

func NewRideGrain(options *RideGrainOptions) *RideGrain {
	return &RideGrain{
		storage:        options.Storage,
		eventPub:       options.EventPub,
		telemetry:      options.Telemetry,
		topic:          options.Topic,
		contextManager: options.ContextManager,
		state: &domain.RideState{
			Status: domain.RideStatusNew,
		},
	}
}

func (g *RideGrain) GetIdentity() *grain.GrainIdentity {
	return g.identity
}

func (g *RideGrain) GetStatus() any {
	return g.state.Status
}

func (g *RideGrain) OnActivate(ctx context.Context, identity *grain.GrainIdentity) error {
	g.identity = identity

	ctx, span := g.traceSpan(ctx, "OnActivate")
	defer span.End()

	// Try to load existing state from storage
	version, err := g.storage.Load(ctx, identity, g.state)
	if err != nil {
		// Check if it's a "not found" error - this is expected for new grains
		if errors.IsNotFound(err) {
			// This is a new grain - initialize with default state
			g.state = &domain.RideState{
				Status: domain.RideStatusNew,
			}
			g.version = 0
			g.telemetry.Logger().InfoContext(ctx, "Activating new ride grain", "identity", identity.String())
			return nil
		}

		// Actual error - failed to load existing grain
		g.telemetry.Logger().ErrorContext(ctx, "Failed to load state for grain",
			"identity", identity.EntityID,
			"error", err)
		return errors.NewTransientErrorf("failed to load ride state: %w", err)
	}

	// Successfully loaded existing grain
	g.version = version
	g.telemetry.Logger().DebugContext(ctx, "Loaded existing ride grain",
		"identity", identity.String(),
		"version", version,
		"status", g.state.Status)
	return nil
}

func (g *RideGrain) OnDeactivate(ctx context.Context) error {
	return nil
}

var messageTypeCache sync.Map

func getMessageType(v any) string {
	t := reflect.TypeOf(v)
	if name, ok := messageTypeCache.Load(t); ok {
		return name.(string)
	}

	name := t.String()
	messageTypeCache.Store(t, name)
	return name
}

func (g *RideGrain) OnReceive(ctx context.Context, msg pkgPorts.Message) (pkgPorts.Message, error) {
	var messageType string
	if m, ok := msg.(pkgPorts.MessageInterface); ok {
		messageType = m.MessageName()
	} else {
		messageType = getMessageType(msg)
	}
	ctx, span := g.traceSpan(ctx, "OnReceive")
	span.SetAttributes(attribute.String("message.type", messageType))
	defer span.End()

	if slices.Contains(terminalStates, g.state.Status) {
		g.telemetry.Logger().WarnContext(ctx, "Received command for ride in terminal state",
			"ride.id", g.identity.EntityID,
			"status", g.state.Status,
			"command.type", messageType)
		span.AddEvent("Grain in terminal state received command")
		return nil, errors.NewBusinessErrorf("cannot process command %T in terminal state %s", msg, g.state.Status)
	}

	g.telemetry.Logger().DebugContext(ctx, "Receiving message", "type", messageType)
	var err error
	var response pkgPorts.Message

	switch message := msg.(type) {
	case *RequestRideCommand:
		response, err = g.handleRequestRide(ctx, message)
	case *CancelRideCommand:
		response, err = g.handleCancelRide(ctx, message)
	case *RideMatchedEvent:
		response, err = g.handleRideMatched(ctx, message)
	case *AcceptRideCommand:
		response, err = g.handleAcceptRide(ctx, message)
	case *RejectRideCommand:
		response, err = g.handleRejectRide(ctx, message)
	case *CompleteRideCommand:
		response, err = g.handleCompleteRide(ctx, message)
	case *StartRideCommand:
		response, err = g.handleStartRide(ctx, message)
	case *RideTimedOutEvent:
		response, err = g.handleRideTimedOut(ctx, message)

	default:
		return nil, errors.NewPermanentErrorf("unhandled message type: %T", msg)
	}

	return response, err
}

func (g *RideGrain) handleRequestRide(ctx context.Context, cmd *RequestRideCommand) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:RequestRide")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return cmd.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.state.Status != domain.RideStatusNew {
				return errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusRequested)
			}
			return nil
		}).
		Step("Transition", func(ctx context.Context) error {
			g.core = &domain.RideCore{
				RiderID:         cmd.RiderID,
				PickupLocation:  cmd.PickupLocation,
				DropoffLocation: cmd.DropoffLocation,
				ServiceType:     cmd.ServiceType,
				Fare:            decimal.NewFromFloat(cmd.Fare),
				Currency:        cmd.Currency,
				RequestID:       cmd.RequestID,
			}
			g.state.Status = domain.RideStatusRequested
			g.version++
			return nil
		}).
		Step("Persist", func(ctx context.Context) error { return g.persist(ctx, cmd) }).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideRequestedEvent{
				RideID:          g.identity.EntityID,
				RequestID:       cmd.RequestID,
				RiderID:         cmd.RiderID,
				PickupLocation:  cmd.PickupLocation,
				DropoffLocation: cmd.DropoffLocation,
				ServiceType:     cmd.ServiceType,
				Fare:            cmd.Fare,
				Currency:        cmd.Currency,
			})
		}).
		End(&RequestRideResponse{RideID: g.identity.EntityID})
}

func (g *RideGrain) handleCancelRide(ctx context.Context, cmd *CancelRideCommand) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:CancelRide")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return cmd.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.core.RiderID != cmd.RiderID {
				return errors.NewBusinessErrorf("only the rider can cancel the ride")
			}
			if g.state.Status != domain.RideStatusRequested && g.state.Status != domain.RideStatusMatched && g.state.Status != domain.RideStatusAccepted {
				return errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusCancelled)
			}
			return nil
		}).
		Step("Transition", func(ctx context.Context) error {
			g.state.Status = domain.RideStatusCancelled
			g.version++
			return nil
		}).
		Step("Persist", func(ctx context.Context) error { return g.persist(ctx, cmd) }).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideCanceledEvent{
				RequestID: cmd.RequestID,
				RiderID:   cmd.RiderID,
				RideID:    cmd.RideID,
			})
		}).
		End(&SuccessResponse{})
}

func (g *RideGrain) handleAcceptRide(ctx context.Context, cmd *AcceptRideCommand) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:AcceptRide")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return cmd.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.state.Status != domain.RideStatusMatched {
				return errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusAccepted)
			}
			return nil
		}).
		Step("Transition", func(ctx context.Context) error {
			g.state.Status = domain.RideStatusAccepted
			g.state.DriverID = &cmd.DriverID
			g.version++
			return nil
		}).
		Step("Persist", func(ctx context.Context) error { return g.persist(ctx, cmd) }).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideAcceptedEvent{
				RequestID: cmd.RequestID,
				DriverID:  cmd.DriverID,
				RideID:    cmd.RideID,
			})
		}).
		End(&SuccessResponse{})
}

func (g *RideGrain) handleRejectRide(ctx context.Context, cmd *RejectRideCommand) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:RejectRide")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return cmd.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.state.Status != domain.RideStatusMatched {
				return errors.NewBusinessErrorf("Cannot reject a ride with state %s", g.state.Status)
			}
			return nil
		}).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideRejectedEvent{
				RequestID: cmd.RequestID,
				DriverID:  cmd.DriverID,
				RideID:    cmd.RideID,
			})
		}).
		End(&SuccessResponse{})
}

func (g *RideGrain) handleStartRide(ctx context.Context, cmd *StartRideCommand) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:StartRide")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return cmd.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.state.Status != domain.RideStatusAccepted {
				return errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusStarted)
			}
			return nil
		}).
		Step("Transition", func(ctx context.Context) error {
			g.state.Status = domain.RideStatusStarted
			g.state.DriverID = &cmd.DriverID
			g.version++
			return nil
		}).
		Step("Persist", func(ctx context.Context) error { return g.persist(ctx, cmd) }).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideStartedEvent{
				RequestID: cmd.RequestID,
				DriverID:  cmd.DriverID,
				RideID:    cmd.RideID,
			})
		}).
		End(&SuccessResponse{})
}

func (g *RideGrain) handleCompleteRide(ctx context.Context, cmd *CompleteRideCommand) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:CompleteRide")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return cmd.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.state.Status != domain.RideStatusStarted {
				return errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusCompleted)
			}
			return nil
		}).
		Step("Transition", func(ctx context.Context) error {
			g.state.Status = domain.RideStatusCompleted
			g.version++
			return nil
		}).
		Step("Persist", func(ctx context.Context) error { return g.persist(ctx, cmd) }).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideCompletedEvent{
				RequestID: cmd.RequestID,
				DriverID:  cmd.DriverID,
				RideID:    cmd.RideID,
			})
		}).
		End(&SuccessResponse{})
}

func (g *RideGrain) handleRideMatched(ctx context.Context, event *RideMatchedEvent) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:RideMatched")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return event.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.state.Status != domain.RideStatusRequested {
				return errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusMatched)
			}
			return nil
		}).
		Step("Transition", func(ctx context.Context) error {
			g.state.Status = domain.RideStatusMatched
			g.state.DriverID = &event.DriverID
			g.version++
			return nil
		}).
		Step("Persist", func(ctx context.Context) error { return g.persist(ctx, event) }).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideMatchedEvent{
				RequestID: event.RequestID,
				DriverID:  event.DriverID,
				RideID:    event.RideID,
			})
		}).
		End(&SuccessResponse{})
}

func (g *RideGrain) handleRideTimedOut(ctx context.Context, event *RideTimedOutEvent) (pkgPorts.Message, error) {
	p := g.Start(ctx, "Handle:RideTimedOut")

	return p.
		Step("ValidateMessage", func(_ context.Context) error { return event.Validate() }).
		Step("ValidateTransition", func(ctx context.Context) error {
			if g.state.Status != domain.RideStatusRequested && g.state.Status != domain.RideStatusMatched {
				return errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusTimedOut)
			}
			return nil
		}).
		Step("Transition", func(ctx context.Context) error {
			g.state.Status = domain.RideStatusTimedOut
			g.version++
			return nil
		}).
		Step("Persist", func(ctx context.Context) error { return g.persist(ctx, event) }).
		Step("Publish", func(ctx context.Context) error {
			return g.publishEvent(ctx, &contracts.RideTimedOutEvent{
				RequestID: event.RequestID,
				RideID:    event.RideID,
			})
		}).
		End(&SuccessResponse{})
}

func (g *RideGrain) publishEvent(ctx context.Context, event contracts.Event) error {
	info, ok := g.contextManager.Extract(ctx)
	if !ok {
		g.telemetry.Logger().WarnContext(ctx, "Failed to extract request info from context, publishing event without trace information",
			"event_type", event.EventType(),
			"ride_id", g.identity.EntityID)
	}
	message := contracts.NewEventMessage(event)
	if ok {
		message.AddHeaders(info.ToByteMap())
	}

	err := g.eventPub.Publish(ctx, contracts.TopicRide, message)
	if err != nil {
		g.telemetry.Logger().ErrorContext(ctx, "Failed to publish event",
			"event_type", event.EventType(),
			"ride_id", g.identity.EntityID,
			"error", err)
		return errors.NewTransientErrorf("failed to publish event %s: %w", event.EventType(), err)
	}
	return nil
}

var spanNameCache = sync.Map{}

func (g *RideGrain) traceSpan(ctx context.Context, method string) (context.Context, trace.Span) {
	name, ok := spanNameCache.Load(method)
	if !ok {
		name = "RideGrain." + method
		spanNameCache.Store(method, name)
	}
	ctx, span := g.telemetry.Tracer().Start(ctx, name.(string),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("grain.kind", g.identity.Kind.String()),
			attribute.String("grain.identity", g.identity.EntityID.String()),
			attribute.String("grain.status", g.state.Status.String()),
			attribute.Int("grain.version", g.version),
		),
	)
	return ctx, span
}

func (g *RideGrain) persist(ctx context.Context, msg pkgPorts.MessageInterface) error {
	data := &domain.GrainData{
		Message:  msg,
		Identity: g.identity,
		Core:     g.core,
		State:    g.state,
		Version:  g.version,
	}
	if err := g.storage.Persist(ctx, g.identity, data); err != nil {
		return errors.NewTransientErrorf("failed to save ride state: %w", err)
	}
	return nil
}
