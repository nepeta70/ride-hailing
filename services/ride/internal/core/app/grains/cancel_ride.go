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

type RejectRideCommand struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (c *RejectRideCommand) CommandName() string {
	return "RejectRide"
}

func (c *RejectRideCommand) Validate() error {
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

type StartRideCommand struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (c *StartRideCommand) CommandName() string {
	return "StartRide"
}

func (c *StartRideCommand) Validate() error {
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

type CompleteRideCommand struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (c *CompleteRideCommand) CommandName() string {
	return "CompleteRide"
}

func (c *CompleteRideCommand) Validate() error {
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
