package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RequestRide struct {
	FareId string
}

func (c *RequestRide) Validate() error {
	if c.FareId == "" {
		return errors.NewValidationErrorf("request information is incomplete")
	}
	return nil
}

type RequestRideHandler struct {
	storage ports.StorageBundle
	logger  pkgPorts.Logger
}

func NewRequestRideHandler(config *config.Config, storage ports.StorageBundle, logger pkgPorts.Logger) *RequestRideHandler {
	return &RequestRideHandler{storage: storage, logger: logger}
}

func (h *RequestRideHandler) Handle(ctx context.Context, cmd RequestRide) (*domain.Ride, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	// TODO: Implement ride request logic
	return nil, nil
}
