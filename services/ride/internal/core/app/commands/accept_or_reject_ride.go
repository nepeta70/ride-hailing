package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
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
	repo ports.RideWriteRepository
}

func NewAcceptOrRejectRideHandler(repo ports.RideWriteRepository) *AcceptOrRejectRideHandler {
	return &AcceptOrRejectRideHandler{repo: repo}
}

func (h *AcceptOrRejectRideHandler) Handle(ctx context.Context, cmd AcceptOrRejectRide) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	// TODO: Implement accept/reject logic
	return nil
}
