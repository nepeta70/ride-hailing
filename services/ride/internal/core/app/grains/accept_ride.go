package grains

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type AcceptRideCommand struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (c *AcceptRideCommand) CommandName() string {
	return "AcceptRide"
}

func (c *AcceptRideCommand) Validate() error {
	if c.RequestID == uuid.Nil {
		return errors.NewValidationErrorf("RequestID cannot be empty")
	}
	if c.DriverID == uuid.Nil {
		return errors.NewValidationErrorf("DriverID cannot be empty")
	}
	if c.RideID == uuid.Nil {
		return errors.NewValidationErrorf("RideID cannot be empty")
	}

	return nil
}
