package grains

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type RequestRideCommand struct {
	RequestID       uuid.UUID
	RiderID         uuid.UUID
	PickupLocation  string
	DropoffLocation string
	ServiceType     string
	Fare            float64
	Currency        string
}

func (c *RequestRideCommand) CommandName() string {
	return "RequestRide"
}

func (c *RequestRideCommand) Validate() error {
	if c.RequestID == uuid.Nil {
		return errors.NewValidationErrorf("RequestID cannot be empty")
	}
	if c.RiderID == uuid.Nil {
		return errors.NewValidationErrorf("RiderID cannot be empty")
	}
	if c.PickupLocation == "" {
		return errors.NewValidationErrorf("PickupLocation cannot be empty")
	}
	if c.DropoffLocation == "" {
		return errors.NewValidationErrorf("DropoffLocation cannot be empty")
	}
	if c.Fare <= 0 {
		return errors.NewValidationErrorf("Fare must be greater than zero")
	}
	if len(c.Currency) != 3 {
		return errors.NewValidationErrorf("Currency must be a 3-letter code")
	}

	return nil
}

type RequestRideResponse struct {
	RideID uuid.UUID
}
