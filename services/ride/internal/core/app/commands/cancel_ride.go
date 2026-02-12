package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type CancelRide struct {
	RideID string
}

func (c *CancelRide) Validate() error {
	if c.RideID == "" {
		return errors.NewValidationErrorf("ride id is required")
	}
	return nil
}

type CancelRideHandler struct {
	silo                 pkgPorts.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	logger               pkgPorts.Logger
}

func NewCancelRideHandler(config *config.Config, storage ports.StorageBundle, grainSystem ports.GrainSystemInterface, logger pkgPorts.Logger) *CancelRideHandler {
	return &CancelRideHandler{
		silo:                 grainSystem.Silo(),
		grainIdentityFactory: grainSystem.GrainIdentityFactory(),
		logger:               logger,
	}
}

func (h *CancelRideHandler) Handle(ctx context.Context, cmd CancelRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement ride cancellation logic
	return nil
}
