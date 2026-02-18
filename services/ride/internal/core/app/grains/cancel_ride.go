package grains

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type CancelRideCommand struct {
	RequestID uuid.UUID
	RiderID   uuid.UUID
	RideID    uuid.UUID
}

func (c *CancelRideCommand) CommandName() string {
	return "CancelRide"
}

func (c *CancelRideCommand) Validate() error {
	if c.RequestID == uuid.Nil {
		return errors.NewValidationErrorf("RequestID cannot be empty")
	}
	if c.RiderID == uuid.Nil {
		return errors.NewValidationErrorf("RiderID cannot be empty")
	}
	if c.RideID == uuid.Nil {
		return errors.NewValidationErrorf("RideID cannot be empty")
	}

	return nil
}
