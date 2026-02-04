package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RequestRide struct {
	RequestID       string
	UserID          string
	PickupLocation  string
	DropoffLocation string
}

func (c *RequestRide) Validate() error {
	if c.RequestID == "" || c.UserID == "" || c.PickupLocation == "" || c.DropoffLocation == "" {
		return errors.NewValidationErrorf("request information is incomplete")
	}
	return nil
}

type RequestRideHandler struct {
	repo ports.RideWriteRepository
}

func NewRequestRideHandler(repo ports.RideWriteRepository) *RequestRideHandler {
	return &RequestRideHandler{repo: repo}
}

func (h *RequestRideHandler) Handle(ctx context.Context, cmd RequestRide) (rideID, driverID, vehicleInfo string, err error) {
	if err := ctx.Err(); err != nil {
		return "", "", "", errors.ErrContextError
	}

	// TODO: Implement ride request logic
	return "", "", "", nil
}
