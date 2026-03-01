package app

import (
	"context"
	"encoding/json"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/grains"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type ApplicationOptions struct {
	Config            *config.Config
	DirectionsService ports.DirectionsService
	Storage           ports.StorageBundle
	GrainSystem       *GrainSystem
	ContextManager    *ctxmgr.ContextManager
	Subscriber        pkgPorts.EventSubscriber
	Telemetry         pkgPorts.TelemetryProvider
}

func (opts *ApplicationOptions) Validate() error {
	if opts.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider is required")
	}
	if opts.Storage == nil {
		return errors.NewValidationErrorf("storage bundle is required")
	}
	if opts.DirectionsService == nil {
		return errors.NewValidationErrorf("directions service is required")
	}
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	if opts.GrainSystem == nil {
		return errors.NewValidationErrorf("grain system is required")
	}
	if opts.Subscriber == nil {
		return errors.NewValidationErrorf("subscriber is required")
	}
	return nil
}

type Application struct {
	Commands          *Commands
	Queries           *Queries
	ContextManager    *ctxmgr.ContextManager
	storage           ports.StorageBundle
	grainSystem       *GrainSystem
	directionsService ports.DirectionsService
	config            *config.Config
	subscriber        pkgPorts.EventSubscriber
	telemetry         pkgPorts.TelemetryProvider
}

type Commands struct {
	EstimateFare   *commands.EstimateFareHandler
	CreateFareRate *commands.CreateFareRateHandler
	UpdateFareRate *commands.UpdateFareRateHandler
	RequestRide    *commands.RequestRideHandler
}

type Queries struct {
	FareRates *queries.GetFareRateHandler
}

func NewApplication(opts *ApplicationOptions) (*Application, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	app := &Application{
		Commands: &Commands{
			EstimateFare:   commands.NewEstimateFareHandler(opts.Config, opts.Telemetry, opts.Storage, opts.DirectionsService),
			CreateFareRate: commands.NewCreateFareRateHandler(opts.Storage.FareRatesWriteRepo()),
			UpdateFareRate: commands.NewUpdateFareRateHandler(opts.Storage.FareRatesWriteRepo()),
			RequestRide:    commands.NewRequestRideHandler(opts.Config, opts.Storage, opts.GrainSystem, opts.Telemetry),
		},
		Queries: &Queries{
			FareRates: queries.NewGetFareRatesHandler(opts.Storage.FareRatesReadRepo()),
		},

		ContextManager:    opts.ContextManager,
		config:            opts.Config,
		storage:           opts.Storage,
		directionsService: opts.DirectionsService,
		grainSystem:       opts.GrainSystem,
		subscriber:        opts.Subscriber,
		telemetry:         opts.Telemetry,
	}
	return app, nil
}

func (a *Application) Start(ctx context.Context) error {
	err := a.subscriber.Subscribe(ctx, contracts.TopicMatching, a.handleRideEvent)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) handleRideEvent(ctx context.Context, headers map[string]string, msg []byte) error {
	ctx = a.telemetry.Propagator().Extract(ctx, propagation.MapCarrier(headers))
	ctx, span := a.telemetry.Tracer().Start(ctx, "Consume Matching Event",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("kafka.topic", contracts.TopicMatching.String()),
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
		span.SetAttributes(attribute.String("error", "missing event-type header"))
		a.telemetry.Logger().ErrorContext(ctx, "Missing event-type header in message")
		return errors.NewValidationErrorf("missing event-type header")
	}
	evt := contracts.EventType(eventType)
	span.SetAttributes(attribute.String("message.type", evt.String()))

	info, ok := ctxmgr.NewInfoFromMap(headers)
	if ok {
		ctx = a.ContextManager.Inject(ctx, info)
	} else {
		span.SetAttributes(attribute.String("error", "failed to create RequestInfo from headers"))
		a.telemetry.Logger().ErrorContext(ctx, "Failed to create RequestInfo from headers")
		return errors.NewValidationErrorf("Failed to create RequestInfo from headers")
	}

	switch evt {
	case contracts.EventTypeRideMatched:
		var payload contracts.RideMatchedEvent
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return errors.NewErrJSONUnmarshal(err)
		}

		ev := &grains.RideMatchedEvent{
			RideID:    payload.RideID,
			DriverID:  payload.DriverID,
			RequestID: info.Trace.RequestID,
		}
		identity := grain.NewGrainIdentity(domain.RideGrainKind, ev.RideID)

		_, err := a.grainSystem.Silo().Ask(ctx, identity, ev)
		if err != nil {
			return err
		}

	default:
	}

	return nil
}
