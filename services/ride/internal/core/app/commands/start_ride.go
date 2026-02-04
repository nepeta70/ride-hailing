package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
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
	repo ports.RideWriteRepository
}

func NewStartRideHandler(repo ports.RideWriteRepository) *StartRideHandler {
	return &StartRideHandler{repo: repo}
}

func (h *StartRideHandler) Handle(ctx context.Context, cmd StartRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement start ride logic
	return nil
}
