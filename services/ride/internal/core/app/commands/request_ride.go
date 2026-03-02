package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/grains"
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

type RequestRideHandler struct {
	fareReadRepo         ports.FareReadRepository
	serviceTypeCache     ports.ServiceTypeCacheInterface
	silo                 pkgPorts.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	telemetry            pkgPorts.TelemetryProvider
}

func NewRequestRideHandler(config *config.Config, storage ports.StorageBundle, grainSystem ports.GrainSystemInterface, telemetry pkgPorts.TelemetryProvider) *RequestRideHandler {
	return &RequestRideHandler{
		fareReadRepo:         storage.FareReadRepo(),
		silo:                 grainSystem.Silo(),
		grainIdentityFactory: grainSystem.GrainIdentityFactory(),
		serviceTypeCache:     storage.ServiceTypeCache(),
		telemetry:            telemetry,
	}
}

func (h *RequestRideHandler) Handle(ctx context.Context, cmd RequestRide) (*grains.RequestRideResponse, error) {
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
	return regResp, nil
}
