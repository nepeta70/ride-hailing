package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type CompleteRide struct {
	RideID string
}

func (c *CompleteRide) Validate() error {
	if c.RideID == "" {
		return errors.NewValidationErrorf("ride id is required")
	}
	return nil
}

type CompleteRideHandler struct {
	silo                 pkgPorts.Silo
	grainIdentityFactory *service.GrainIdentityFactory
	logger               pkgPorts.Logger
}

func NewCompleteRideHandler(config *config.Config, storage ports.StorageBundle, grainSystem ports.GrainSystemInterface, logger pkgPorts.Logger) *CompleteRideHandler {
	return &CompleteRideHandler{
		silo:                 grainSystem.Silo(),
		grainIdentityFactory: grainSystem.GrainIdentityFactory(),
		logger:               logger,
	}
}

func (h *CompleteRideHandler) Handle(ctx context.Context, cmd CompleteRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement complete ride logic
	return nil
}
