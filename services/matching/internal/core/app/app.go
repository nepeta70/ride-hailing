package app

import (
	"context"
	"encoding/json"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/matching/internal/core/service"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type AppOptions struct {
	Service        *service.MatchingService
	Subscriber     ports.EventSubscriber
	EventPublisher ports.EventPublisher
	ContextManager *ctxmgr.ContextManager
	Telemetry      ports.TelemetryProvider
}

func (o *AppOptions) Validate() error {
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("TelemetryProvider is required")
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
	service        *service.MatchingService
	subscriber     ports.EventSubscriber
	publisher      ports.EventPublisher
	contextManager *ctxmgr.ContextManager
	telemetry      ports.TelemetryProvider
}

func NewApplication(options *AppOptions) (*Application, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	return &Application{
		telemetry:      options.Telemetry,
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
	ctx = a.telemetry.Propagator().Extract(ctx, propagation.MapCarrier(headers))
	ctx, span := a.telemetry.Tracer().Start(ctx, "Consume Ride Event",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("topic", contracts.TopicRide.String()),
			attribute.String("operation", "consume"),
			attribute.String("message_type", "RideEvent"),
		),
	)

	defer span.End()

	var event contracts.EventMessage
	if err := json.Unmarshal(msg, &event); err != nil {
		a.telemetry.Logger().ErrorContext(ctx, "Poison message received", "error", err)
		return errors.NewErrJSONUnmarshal(err)
	}

	info, ok := ctxmgr.NewInfoFromMap(headers)
	if !ok {
		a.telemetry.Logger().ErrorContext(ctx, "Failed to create RequestInfo from headers")
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

		a.telemetry.Logger().DebugContext(ctx, "Received RideRequestedEvent", "ride_id", rideEvent.RideID.String())
		_, err := a.service.MatchRiderToDriver(ctx, headers, rideEvent)
		if err != nil {
			a.telemetry.Logger().ErrorContext(ctx, "Error matching rider to driver", "error", err)
			return err
		}

	default:
	}

	return nil
}
