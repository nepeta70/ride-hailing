package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/app/grains"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RequestRide struct {
	FareID      string
	RiderID     string
	RequestID   string
	ServiceType string
}

func (c *RequestRide) Validate() error {
	err := uuid.Validate(c.RiderID)
	if err != nil {
		return errors.NewValidationErrorf("invalid rider ID format")
	}

	if c.FareID == "" {
		return errors.NewValidationErrorf("request information is incomplete")
	}
	err = uuid.Validate(c.FareID)
	if err != nil {
		return errors.NewValidationErrorf("invalid fare ID format")
	}

	err = uuid.Validate(c.RequestID)
	if err != nil {
		return errors.NewValidationErrorf("invalid request ID format")
	}

	if c.ServiceType == "" {
		return errors.NewValidationErrorf("service type is required")
	}
	return nil
}

type RequestRideHandler struct {
	fareReadRepo         ports.FareReadRepository
	silo                 pkgPorts.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	logger               pkgPorts.Logger
}

func NewRequestRideHandler(config *config.Config, storage ports.StorageBundle, grainSystem ports.GrainSystemInterface, logger pkgPorts.Logger) *RequestRideHandler {
	return &RequestRideHandler{
		fareReadRepo:         storage.FareReadRepo(),
		silo:                 grainSystem.Silo(),
		grainIdentityFactory: grainSystem.GrainIdentityFactory(),
		logger:               logger,
	}
}

func (h *RequestRideHandler) Handle(ctx context.Context, cmd RequestRide) (*grains.RequestRideResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	err := cmd.Validate()
	if err != nil {
		return nil, err
	}

	fareID := uuid.MustParse(cmd.FareID)
	fare, err := h.fareReadRepo.GetByID(ctx, fareID)
	if err != nil {
		return nil, err
	}

	if fare == nil {
		return nil, errors.NewErrNotFound("fare not found")
	}
	// TODO: servicetype validation
	riderId := uuid.MustParse(cmd.RiderID)
	requestId := uuid.MustParse(cmd.RequestID)
	identity := h.grainIdentityFactory.NewRideGrainIdentity(riderId, requestId, fareID)
	command := &grains.RequestRideCommand{
		RequestID:       requestId,
		RiderID:         riderId,
		PickupLocation:  fare.PickupLocation,
		DropoffLocation: fare.DropoffLocation,
		ServiceType:     cmd.ServiceType,
		Fare:            fare.Fares[cmd.ServiceType], // TODO: validate it
		Currency:        fare.Currency,
	}

	resp, err := h.silo.Ask(ctx, identity, command)
	if err != nil {
		return nil, err
	}

	regResp := resp.(*grains.RequestRideResponse)
	return regResp, nil
}
