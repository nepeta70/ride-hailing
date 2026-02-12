package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type StartRide struct {
	RideID string
}

func (c *StartRide) Validate() error {
	if c.RideID == "" {
		return errors.NewValidationErrorf("ride id is required")
	}
	return nil
}

type StartRideHandler struct {
	silo                 pkgPorts.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	logger               pkgPorts.Logger
}

func NewStartRideHandler(config *config.Config, storage ports.StorageBundle, grainSystem ports.GrainSystemInterface, logger pkgPorts.Logger) *StartRideHandler {
	return &StartRideHandler{
		silo:                 grainSystem.Silo(),
		grainIdentityFactory: grainSystem.GrainIdentityFactory(),
		logger:               logger,
	}
}

func (h *StartRideHandler) Handle(ctx context.Context, cmd StartRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement start ride logic
	return nil
}
