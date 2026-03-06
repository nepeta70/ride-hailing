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
	"go.opentelemetry.io/otel/codes"
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
	ctx = a.telemetry.Propagator().Extract(context.Background(), propagation.MapCarrier(headers))
	ctx, span := a.telemetry.Tracer().Start(ctx, "Consume Ride Event",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("kafka.topic", contracts.TopicRide.String()),
			attribute.String("kafka.operation", "consume"),
		),
	)

	defer span.End()

	var envelope struct {
		Payload json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(msg, &envelope); err != nil {
		a.telemetry.Logger().ErrorContext(ctx, "Poison message received", "error", err)
		return errors.NewErrJSONUnmarshal(err)
	}

	// Read event type from headers
	eventType, ok := headers["event-type"]
	if !ok || eventType == "" {
		return errors.NewValidationErrorf("missing event-type header")
	}
	evt := contracts.EventType(eventType)
	span.SetAttributes(attribute.String("message.type", evt.String()))

	info, ok := ctxmgr.NewInfoFromMap(headers)
	if !ok {
		a.telemetry.Logger().ErrorContext(ctx, "Failed to create RequestInfo from headers")
		return errors.NewValidationErrorf("Failed to create RequestInfo from headers")
	}
	ctx = a.contextManager.Inject(ctx, info)
	switch evt {
	case contracts.EventTypeRideRequested:
		span.SetAttributes(attribute.Bool("consumed", true))
		var payload contracts.RideRequestedEvent
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return errors.NewErrJSONUnmarshal(err)
		}

		a.telemetry.Logger().DebugContext(ctx, "Received RideRequestedEvent", "ride_id", payload.RideID.String())
		a.service.MatchRide(ctx, headers, &payload)

	case contracts.EventTypeRideCanceled:
		span.SetAttributes(attribute.Bool("consumed", true))
		var payload contracts.RideCanceledEvent
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return errors.NewErrJSONUnmarshal(err)
		}

		a.telemetry.Logger().DebugContext(ctx, "Received RideCanceledEvent", "ride_id", payload.RideID.String())
		a.service.HandleCancelRide(ctx, &payload)

	default:
		span.SetAttributes(attribute.Bool("skipped", true))
	}
	span.SetStatus(codes.Ok, "message processed successfully")
	return nil
}
