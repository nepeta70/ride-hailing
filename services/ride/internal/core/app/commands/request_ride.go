package commands

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/grains"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RequestRide struct {
	FareID      uuid.UUID
	RiderID     uuid.UUID
	RequestID   uuid.UUID
	ServiceType string
}

func (c *RequestRide) Validate() error {
	if c.FareID == uuid.Nil {
		return errors.NewValidationErrorf("invalid fare ID format")
	}

	if c.ServiceType == "" {
		return errors.NewValidationErrorf("service type is required")
	}
	return nil
}

type RequestRideHandlerOpts struct {
	ContextManager       *ctxmgr.ContextManager
	Config               *config.Config
	FareReadRepo         ports.FareReadRepository
	Silo                 pkgPorts.Silo
	GrainIdentityFactory *service.GrainIdentityFactory
	Telemetry            pkgPorts.TelemetryProvider
	ServiceTypeCache     ports.ServiceTypeCacheInterface
}

func (opts *RequestRideHandlerOpts) Validate() error {
	if opts.ContextManager == nil {
		return errors.NewValidationErrorf("context manager cannot be nil")
	}
	if opts.Config == nil {
		return errors.NewValidationErrorf("config cannot be nil")
	}
	if opts.FareReadRepo == nil {
		return errors.NewValidationErrorf("fare read repository cannot be nil")
	}
	if opts.Silo == nil {
		return errors.NewValidationErrorf("silo cannot be nil")
	}
	if opts.GrainIdentityFactory == nil {
		return errors.NewValidationErrorf("grain identity factory cannot be nil")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry provider cannot be nil")
	}
	if opts.ServiceTypeCache == nil {
		return errors.NewValidationErrorf("service type cache cannot be nil")
	}
	return nil
}

type RequestRideHandler struct {
	config               *config.Config
	fareReadRepo         ports.FareReadRepository
	serviceTypeCache     ports.ServiceTypeCacheInterface
	silo                 pkgPorts.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	telemetry            pkgPorts.TelemetryProvider
	contextManager       ctxmgr.ContextManager
}

func NewRequestRideHandler(opts *RequestRideHandlerOpts) *RequestRideHandler {
	if err := opts.Validate(); err != nil {
		panic(err)
	}

	return &RequestRideHandler{
		config:               opts.Config,
		fareReadRepo:         opts.FareReadRepo,
		silo:                 opts.Silo,
		grainIdentityFactory: opts.GrainIdentityFactory,
		serviceTypeCache:     opts.ServiceTypeCache,
		telemetry:            opts.Telemetry,
		contextManager:       *opts.ContextManager,
	}
}

func (h *RequestRideHandler) Handle(ctx context.Context, cmd *RequestRide) (*grains.RequestRideResponse, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "RequestRideHandler.Handle",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("handler", "RequestRideHandler"),
			attribute.String("method", "Handle"),
			attribute.String("request.id", cmd.RequestID.String()),
			attribute.String("fare.id", cmd.FareID.String()),
			attribute.String("service.type", cmd.ServiceType),
			attribute.String("rider.id", cmd.RiderID.String()),
		),
	)
	defer span.End()
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	err := cmd.Validate()
	if err != nil {
		return nil, err
	}

	fare, err := h.fareReadRepo.GetByID(ctx, cmd.FareID)
	if err != nil {
		return nil, err
	}

	if fare == nil {
		return nil, errors.NewErrNotFound("fare not found")
	}

	_, ok := h.serviceTypeCache.GetServiceTypeByCode(ctx, cmd.ServiceType)
	if !ok {
		return nil, errors.NewErrNotFound("service type not found")
	}

	identity := h.grainIdentityFactory.NewRideGrainIdentity(cmd.RiderID, cmd.RequestID, cmd.FareID)
	command := &grains.RequestRideCommand{
		RequestID:       cmd.RequestID,
		RiderID:         cmd.RiderID,
		PickupLocation:  fare.PickupLocation,
		DropoffLocation: fare.DropoffLocation,
		ServiceType:     cmd.ServiceType,
		Fare:            fare.Fares[cmd.ServiceType],
		Currency:        fare.Currency,
	}

	resp, err := h.silo.Ask(ctx, identity, command)
	if err != nil {
		return nil, err
	}

	regResp := resp.(*grains.RequestRideResponse)
	requestInfo, _ := h.contextManager.Extract(ctx)

	spanCtx := trace.SpanContextFromContext(ctx)

	go h.startTimeoutMonitor(spanCtx, requestInfo, regResp.RideID)

	return regResp, nil
}

func (h *RequestRideHandler) startTimeoutMonitor(
	parentSpan trace.SpanContext,
	info *ctxmgr.RequestInfo,
	rideID uuid.UUID) {

	time.Sleep(h.config.RideConfig.RideRequestTimeout)

	// 1. Create a fresh context that inherits the Trace ID
	// This ensures the Timeout shows up in the same Jaeger/Honeycomb trace.
	ctx := trace.ContextWithSpanContext(context.Background(), parentSpan)

	// 2. Re-inject the Request Metadata (API Keys, UserID, etc.)
	if info != nil {
		ctx = h.contextManager.Inject(ctx, info)
	}

	// Start a new span for the timeout trigger itself
	ctx, span := h.telemetry.Tracer().Start(ctx, "RequestRideHandler.TimeoutTrigger",
		trace.WithAttributes(
			attribute.String("ride.id", rideID.String()),
			attribute.String("reason", "automatic_matching_timeout"),
		),
	)
	defer span.End()

	identity := grain.NewGrainIdentity(domain.RideGrainKind, rideID)

	// 3. Execute the Timeout Command
	// The Grain will check its status and transition if necessary.
	g, err := h.silo.GetOrActivate(ctx, identity)

	if err != nil {
		span.RecordError(err)
		h.telemetry.Logger().ErrorContext(ctx, "Failed to activate grain for timeout",
			"ride.id", rideID,
			"error", err)
		return
	}

	if g.GetStatus() == domain.RideStatusRequested {
		timeoutEvent := &grains.RideTimedOutEvent{
			RequestID: info.Trace.RequestID,
			RideID:    rideID,
		}
		resp, err := h.silo.Ask(ctx, identity, timeoutEvent)
		if err != nil {
			span.RecordError(err)
			h.telemetry.Logger().ErrorContext(ctx, "Failed to trigger matching timeout",
				"ride.id", rideID,
				"error", err)
		}

		span.AddEvent("Matching timeout triggered", trace.WithAttributes(
			attribute.String("ride.id", rideID.String()),
			attribute.String("response", resp.(string)),
		))
	}
}
