package grains

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"

	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RideGrain struct {
	identity *grain.GrainIdentity
	core     *domain.RideCore
	state    *domain.RideState
	version  int
	storage  ports.GrainStorage
	eventPub pkgPorts.EventPublisher
	logger   pkgPorts.Logger
	topic    contracts.Topic
}

var _ pkgPorts.Grain = (*RideGrain)(nil)

func NewRideGrain(storage ports.GrainStorage, eventPub pkgPorts.EventPublisher, logger pkgPorts.Logger, topic contracts.Topic) *RideGrain {
	return &RideGrain{
		storage:  storage,
		eventPub: eventPub,
		logger:   logger,
		topic:    topic,
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
	switch cmd := msg.(type) {
	case *RequestRideCommand:
		return g.handleRequestRide(ctx, cmd)
	// case *ReserveSeatsCommand:
	// 	return g.handleReserveSeats(ctx, cmd)
	// case *ReleaseSeatsCommand:
	// 	return g.handleReleaseSeats(ctx, cmd)
	// case *GetCarStateQuery:
	// 	return g.handleGetState(ctx, cmd)
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

	return &RequestRideResponse{RideID: g.identity.EntityID}, nil
}
