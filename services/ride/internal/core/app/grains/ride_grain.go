package grains

import (
	"context"
	"fmt"
	"slices"

	"github.com/shopspring/decimal"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"

	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RideGrainOptions struct {
	Storage        ports.GrainStorage
	EventPub       pkgPorts.EventPublisher
	Logger         pkgPorts.Logger
	Topic          contracts.Topic
	ContextManager *ctxmgr.ContextManager
}

func (opts *RideGrainOptions) Validate() error {
	if opts.Storage == nil {
		return errors.NewValidationErrorf("grain storage is required")
	}
	if opts.EventPub == nil {
		return errors.NewValidationErrorf("event publisher is required")
	}
	if opts.Logger == nil {
		return errors.NewValidationErrorf("logger is required")
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
	logger         pkgPorts.Logger
	topic          contracts.Topic
	contextManager *ctxmgr.ContextManager
}

var _ pkgPorts.Grain = (*RideGrain)(nil)

func NewRideGrain(options *RideGrainOptions) *RideGrain {
	return &RideGrain{
		storage:        options.Storage,
		eventPub:       options.EventPub,
		logger:         options.Logger,
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

func (g *RideGrain) OnActivate(ctx context.Context, identity *grain.GrainIdentity) error {
	g.identity = identity

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
			g.logger.Info("Activating new ride grain", "identity", identity.String())
			return nil
		}

		// Actual error - failed to load existing grain
		g.logger.Error("Failed to load state for grain",
			"identity", identity.EntityID,
			"error", err)
		return errors.NewTransientErrorf("failed to load ride state: %w", err)
	}

	// Successfully loaded existing grain
	g.version = version
	g.logger.Debug("Loaded existing ride grain",
		"identity", identity.String(),
		"version", version,
		"status", g.state.Status)
	return nil
}

func (g *RideGrain) OnDeactivate(ctx context.Context) error {
	return nil
}

func (g *RideGrain) OnReceive(ctx context.Context, msg pkgPorts.Message) (pkgPorts.Message, error) {
	messageType := fmt.Sprintf("%T", msg)
	if slices.Contains(terminalStates, g.state.Status) {
		g.logger.Warn("Received command for ride in terminal state",
			"ride_id", g.identity.EntityID,
			"status", g.state.Status,
			"command_type", messageType)
		return nil, errors.NewBusinessErrorf("cannot process command %T in terminal state %s", msg, g.state.Status)
	}

	g.logger.Debug("Receiving message", "type", messageType)
	switch cmd := msg.(type) {
	case *RequestRideCommand:
		return g.handleRequestRide(ctx, cmd)
	case *CancelRideCommand:
		return g.handleCancelRide(ctx, cmd)
	case *AcceptRideCommand:
		return g.handleAcceptRide(ctx, cmd)
	case *RejectRideCommand:
		return g.handleRejectRide(ctx, cmd)
	case *CompleteRideCommand:
		return g.handleCompleteRide(ctx, cmd)
	case *StartRideCommand:
		return g.handleStartRide(ctx, cmd)

	default:
		return nil, errors.NewPermanentErrorf("unhandled message type: %T", msg)
	}
}

func (g *RideGrain) handleRequestRide(ctx context.Context, cmd *RequestRideCommand) (pkgPorts.Message, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	if g.state.Status != domain.RideStatusNew {
		return nil, errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusRequested)
	}

	g.core = &domain.RideCore{
		RequestID:       cmd.RequestID,
		RiderID:         cmd.RiderID,
		PickupLocation:  cmd.PickupLocation,
		DropoffLocation: cmd.DropoffLocation,
		ServiceType:     cmd.ServiceType,
		Fare:            decimal.NewFromFloat(cmd.Fare),
		Currency:        cmd.Currency,
	}
	g.state.Status = domain.RideStatusRequested
	g.version++

	// Persist state and publish event
	data := &domain.GrainData{
		Command:  cmd,
		Identity: g.identity,
		Core:     g.core,
		State:    g.state,
		Version:  g.version,
	}
	if err := g.storage.Persist(ctx, g.identity, data); err != nil {
		return nil, errors.NewTransientErrorf("failed to save ride state: %w", err)
	}

	event := &contracts.RideRequestedEvent{
		RideID:          g.identity.EntityID,
		RequestID:       cmd.RequestID,
		RiderID:         cmd.RiderID,
		PickupLocation:  cmd.PickupLocation,
		DropoffLocation: cmd.DropoffLocation,
		ServiceType:     cmd.ServiceType,
		Fare:            cmd.Fare,
		Currency:        cmd.Currency,
	}
	err := g.publishEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	return &RequestRideResponse{RideID: g.identity.EntityID}, nil
}

func (g *RideGrain) handleCancelRide(ctx context.Context, cmd *CancelRideCommand) (pkgPorts.Message, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	if g.core.RiderID != cmd.RiderID {
		return nil, errors.NewBusinessErrorf("only the rider can cancel the ride")
	}
	if g.state.Status != domain.RideStatusRequested && g.state.Status != domain.RideStatusAccepted {
		return nil, errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusCancelled)
	}

	g.state.Status = domain.RideStatusCancelled
	g.version++

	// Persist state and publish event
	data := &domain.GrainData{
		Command:  cmd,
		Identity: g.identity,
		Core:     g.core,
		State:    g.state,
		Version:  g.version,
	}
	if err := g.storage.Persist(ctx, g.identity, data); err != nil {
		return nil, errors.NewTransientErrorf("failed to save ride state: %w", err)
	}

	event := &contracts.RideCanceledEvent{
		RequestID: cmd.RequestID,
		RiderID:   cmd.RiderID,
		RideID:    cmd.RideID,
	}
	err := g.publishEvent(ctx, event)
	if err != nil {
		return nil, err
	}
	return &SuccessResponse{}, nil
}

func (g *RideGrain) handleAcceptRide(ctx context.Context, cmd *AcceptRideCommand) (pkgPorts.Message, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	if g.state.Status != domain.RideStatusRequested {
		return nil, errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusAccepted)
	}

	g.state.Status = domain.RideStatusAccepted
	g.state.DriverID = &cmd.DriverID
	g.version++

	// Persist state and publish event
	data := &domain.GrainData{
		Command:  cmd,
		Identity: g.identity,
		Core:     g.core,
		State:    g.state,
		Version:  g.version,
	}
	if err := g.storage.Persist(ctx, g.identity, data); err != nil {
		return nil, errors.NewTransientErrorf("failed to save ride state: %w", err)
	}

	event := &contracts.RideAcceptedEvent{
		RequestID: cmd.RequestID,
		DriverID:  cmd.DriverID,
		RideID:    cmd.RideID,
	}
	err := g.publishEvent(ctx, event)
	if err != nil {
		return nil, err
	}
	return &SuccessResponse{}, nil
}

func (g *RideGrain) handleRejectRide(ctx context.Context, cmd *RejectRideCommand) (pkgPorts.Message, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	if g.state.Status != domain.RideStatusRequested {
		return nil, errors.NewBusinessErrorf("Cannot reject a ride with state %s", g.state.Status)
	}

	event := &contracts.RideRejectedEvent{
		RequestID: cmd.RequestID,
		DriverID:  cmd.DriverID,
		RideID:    cmd.RideID,
	}
	err := g.publishEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{}, nil
}

func (g *RideGrain) handleStartRide(ctx context.Context, cmd *StartRideCommand) (pkgPorts.Message, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	if g.state.Status != domain.RideStatusAccepted {
		return nil, errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusStarted)
	}

	g.state.Status = domain.RideStatusStarted
	g.version++

	// Persist state and publish event
	data := &domain.GrainData{
		Command:  cmd,
		Identity: g.identity,
		Core:     g.core,
		State:    g.state,
		Version:  g.version,
	}
	if err := g.storage.Persist(ctx, g.identity, data); err != nil {
		return nil, errors.NewTransientErrorf("failed to save ride state: %w", err)
	}

	event := &contracts.RideStartedEvent{
		RequestID: cmd.RequestID,
		DriverID:  cmd.DriverID,
		RideID:    cmd.RideID,
	}
	err := g.publishEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{}, nil
}

func (g *RideGrain) handleCompleteRide(ctx context.Context, cmd *CompleteRideCommand) (pkgPorts.Message, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	if g.state.Status != domain.RideStatusStarted {
		return nil, errors.NewBusinessErrorf("invalid ride status transition %s to %s", g.state.Status, domain.RideStatusCompleted)
	}

	g.state.Status = domain.RideStatusCompleted
	g.version++

	// Persist state and publish event
	data := &domain.GrainData{
		Command:  cmd,
		Identity: g.identity,
		Core:     g.core,
		State:    g.state,
		Version:  g.version,
	}
	if err := g.storage.Persist(ctx, g.identity, data); err != nil {
		return nil, errors.NewTransientErrorf("failed to save ride state: %w", err)
	}

	event := &contracts.RideCompletedEvent{
		RequestID: cmd.RequestID,
		DriverID:  cmd.DriverID,
		RideID:    cmd.RideID,
	}
	err := g.publishEvent(ctx, event)
	if err != nil {
		return nil, err
	}
	return &SuccessResponse{}, nil
}

func (g *RideGrain) publishEvent(ctx context.Context, event contracts.Event) error {
	info, ok := g.contextManager.Extract(ctx)
	if !ok {
		g.logger.Warn("Failed to extract request info from context, publishing event without trace information",
			"event_type", event.EventType(),
			"ride_id", g.identity.EntityID)
	}
	message := contracts.NewEventMessage(event)
	message.AddHeaders(info.ToByteMap())
	err := g.eventPub.Publish(ctx, contracts.TopicRide, message)
	if err != nil {
		g.logger.Error("Failed to publish event",
			"event_type", event.EventType(),
			"ride_id", g.identity.EntityID,
			"error", err)
		return errors.NewTransientErrorf("failed to publish event %s: %w", event.EventType(), err)
	}
	return nil
}
