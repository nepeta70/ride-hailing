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
			EstimateFare:   commands.NewEstimateFareHandler(opts.Config, opts.Telemetry.Logger(), opts.Storage, opts.DirectionsService),
			CreateFareRate: commands.NewCreateFareRateHandler(opts.Storage.FareRatesWriteRepo()),
			UpdateFareRate: commands.NewUpdateFareRateHandler(opts.Storage.FareRatesWriteRepo()),
			RequestRide:    commands.NewRequestRideHandler(opts.Config, opts.Storage, opts.GrainSystem, opts.Telemetry.Logger()),
		},
		Queries: &Queries{
			FareRates: queries.NewGetFareRatesHandler(nil), // TODO
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

	var event contracts.EventMessage
	if err := json.Unmarshal(msg, &event); err != nil {
		a.telemetry.Logger().ErrorContext(ctx, "Poison message received", "error", err)
		return errors.NewErrJSONUnmarshal(err)
	}
	span.SetAttributes(attribute.String("message.type", event.EventType.String()))

	info, ok := ctxmgr.NewInfoFromMap(headers)
	if !ok {
		a.telemetry.Logger().ErrorContext(ctx, "Failed to create RequestInfo from headers")
		return errors.NewValidationErrorf("Failed to create RequestInfo from headers")
	}
	a.ContextManager.Inject(ctx, info)
	switch event.EventType {
	case contracts.EventTypeRideMatched:
		var payload contracts.RideMatchedEvent
		payloadBytes, _ := json.Marshal(event.Payload)
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return errors.NewErrJSONUnmarshal(err)
		}
		rideEvent := &payload

		cmd := &grains.CompleteRideCommand{
			RideID:    rideEvent.RideID,
			DriverID:  rideEvent.DriverID,
			RequestID: info.Trace.RequestID,
		}
		identity := grain.NewGrainIdentity(domain.RideGrainKind, cmd.RideID)

		_, err := a.grainSystem.Silo().Ask(ctx, identity, cmd)
		if err != nil {
			return err
		}

	default:
	}

	return nil
}
