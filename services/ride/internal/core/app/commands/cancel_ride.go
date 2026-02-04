package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
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
	repo ports.RideWriteRepository
}

func NewCancelRideHandler(repo ports.RideWriteRepository) *CancelRideHandler {
	return &CancelRideHandler{repo: repo}
}

func (h *CancelRideHandler) Handle(ctx context.Context, cmd CancelRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement ride cancellation logic
	return nil
}
