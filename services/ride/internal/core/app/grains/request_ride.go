package grains

import (
	"github.com/google/uuid"
	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RequestRideCommand struct {
	RequestID       uuid.UUID
	RiderID         uuid.UUID
	PickupLocation  *core.Coordinates
	DropoffLocation *core.Coordinates
	ServiceType     string
	Fare            float64
	Currency        string
}

func (c *RequestRideCommand) MessageName() string {
	return "RequestRide"
}

func (c *RequestRideCommand) Validate() error {
	if c.RequestID == uuid.Nil {
		return errors.NewValidationErrorf("RequestID cannot be empty")
	}
	if c.RiderID == uuid.Nil {
		return errors.NewValidationErrorf("RiderID cannot be empty")
	}
	err := c.PickupLocation.Validate()
	if err != nil {
		return errors.NewValidationErrorf("invalid pickup location: %v", err)
	}
	err = c.DropoffLocation.Validate()
	if err != nil {
		return errors.NewValidationErrorf("invalid dropoff location: %v", err)
	}
	if c.Fare <= 0 {
		return errors.NewValidationErrorf("Fare must be greater than zero")
	}
	if len(c.Currency) != 3 {
		return errors.NewValidationErrorf("Currency must be a 3-letter code")
	}

	return nil
}

var _ ports.MessageInterface = (*RequestRideCommand)(nil)

type RequestRideResponse struct {
	RideID uuid.UUID
}
