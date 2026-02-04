package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
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
	repo ports.RideWriteRepository
}

func NewCompleteRideHandler(repo ports.RideWriteRepository) *CompleteRideHandler {
	return &CompleteRideHandler{repo: repo}
}

func (h *CompleteRideHandler) Handle(ctx context.Context, cmd CompleteRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement complete ride logic
	return nil
}
