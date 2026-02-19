package app

import (
	"context"
	"encoding/json"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/matching/internal/core/service"
)

type AppOptions struct {
	Logger         ports.Logger
	Metrics        ports.Metrics
	Service        *service.MatchingService
	Subscriber     ports.EventSubscriber
	EventPublisher ports.EventPublisher
}

func (o *AppOptions) Validate() error {
	if o.Logger == nil {
		return errors.NewValidationErrorf("Logger is required")
	}
	if o.Metrics == nil {
		return errors.NewValidationErrorf("Metrics is required")
	}
	if o.Service == nil {
		return errors.NewValidationErrorf("MatchingService is required")
	}
	if o.Subscriber == nil {
		return errors.NewValidationErrorf("Subscriber is required")
	}
	if o.EventPublisher == nil {
		return errors.NewValidationErrorf("EventPublisher is required")
	}
	return nil
}

type Application struct {
	logger     ports.Logger
	metrics    ports.Metrics
	service    *service.MatchingService
	subscriber ports.EventSubscriber
	publisher  ports.EventPublisher
}

func NewApplication(options *AppOptions) (*Application, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	return &Application{
		logger:     options.Logger,
		metrics:    options.Metrics,
		service:    options.Service,
		subscriber: options.Subscriber,
		publisher:  options.EventPublisher,
	}, nil
}

func (a *Application) Start(ctx context.Context) error {
	err := a.subscriber.Subscribe(ctx, contracts.TopicRide, a.handleRideEvent)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) handleRideEvent(ctx context.Context, msg []byte) error {
	var message contracts.EventMessage
	if err := json.Unmarshal(msg, &message); err != nil {
		a.logger.Error("CRITICAL: Poison message received", "error", err)
		return errors.NewErrJSONUnmarshal(err)
	}

	switch message.EventType {
	case contracts.EventTypeRideRequested:
		if rideEvent, ok := message.Payload.(*contracts.RideRequestedEvent); ok {
			a.logger.Debug("Received RideRequestedEvent, adding to waitlist", "ride_id", rideEvent.RideID)
			// TODO: match with driver
		}

	default:
	}

	return nil
}
