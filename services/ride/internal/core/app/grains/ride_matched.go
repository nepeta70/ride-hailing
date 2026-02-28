package grains

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RideMatchedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideMatchedEvent) MessageName() string {
	return "RideMatched"
}

func (e *RideMatchedEvent) Validate() error {
	if e.RequestID == uuid.Nil {
		return errors.NewValidationErrorf("RequestID cannot be empty")
	}
	if e.DriverID == uuid.Nil {
		return errors.NewValidationErrorf("DriverID cannot be empty")
	}
	if e.RideID == uuid.Nil {
		return errors.NewValidationErrorf("RideID cannot be empty")
	}

	return nil
}

var _ ports.MessageInterface = (*RideMatchedEvent)(nil)
