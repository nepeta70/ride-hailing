package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type AcceptOrRejectRide struct {
	RideID   string
	DriverID string
	Accept   bool
}

func (c *AcceptOrRejectRide) Validate() error {
	if c.RideID == "" || c.DriverID == "" {
		return errors.NewValidationErrorf("ride id and driver id are required")
	}
	return nil
}

type AcceptOrRejectRideHandler struct {
	silo                 pkgPorts.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	logger               pkgPorts.Logger
}

func NewAcceptOrRejectRideHandler(config *config.Config, storage ports.StorageBundle, grainSystem ports.GrainSystemInterface, logger pkgPorts.Logger) *AcceptOrRejectRideHandler {
	return &AcceptOrRejectRideHandler{
		silo:                 grainSystem.Silo(),
		grainIdentityFactory: grainSystem.GrainIdentityFactory(),
		logger:               logger,
	}
}

func (h *AcceptOrRejectRideHandler) Handle(ctx context.Context, cmd AcceptOrRejectRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement accept/reject logic
	return nil
}
