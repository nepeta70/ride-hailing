package app

import (
	"context"
	"encoding/json"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
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
	ContextManager *ctxmgr.ContextManager
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
	if o.ContextManager == nil {
		return errors.NewValidationErrorf("ContextManager is required")
	}
	return nil
}

type Application struct {
	logger         ports.Logger
	metrics        ports.Metrics
	service        *service.MatchingService
	subscriber     ports.EventSubscriber
	publisher      ports.EventPublisher
	contextManager *ctxmgr.ContextManager
}

func NewApplication(options *AppOptions) (*Application, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	return &Application{
		logger:         options.Logger,
		metrics:        options.Metrics,
		service:        options.Service,
		subscriber:     options.Subscriber,
		publisher:      options.EventPublisher,
		contextManager: options.ContextManager,
	}, nil
}

func (a *Application) Start(ctx context.Context) error {
	err := a.subscriber.Subscribe(ctx, contracts.TopicRide, a.handleRideEvent)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) handleRideEvent(ctx context.Context, headers map[string]string, msg []byte) error {
	var event contracts.EventMessage
	if err := json.Unmarshal(msg, &event); err != nil {
		a.logger.Error("Poison message received", "error", err)
		return errors.NewErrJSONUnmarshal(err)
	}

	info, ok := ctxmgr.NewInfoFromMap(headers)
	if !ok {
		a.logger.Error("Failed to create RequestInfo from headers")
		return errors.NewValidationErrorf("Failed to create RequestInfo from headers")
	}
	a.contextManager.Inject(ctx, info)
	switch event.EventType {
	case contracts.EventTypeRideRequested:
		var payload contracts.RideRequestedEvent
		payloadBytes, _ := json.Marshal(event.Payload)
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return errors.NewErrJSONUnmarshal(err)
		}
		rideEvent := &payload

		a.logger.Debug("Received RideRequestedEvent", "ride_id", rideEvent.RideID.String())
		_, err := a.service.MatchRiderToDriver(ctx, headers, rideEvent)
		if err != nil {
			a.logger.Error("Error matching rider to driver", "error", err)
			return err
		}

	default:
	}

	return nil
}
