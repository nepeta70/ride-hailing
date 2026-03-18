package grains

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RideTimedOutEvent struct {
	RequestID uuid.UUID
	RideID    uuid.UUID
}

func (e *RideTimedOutEvent) MessageName() string {
	return "RideTimedOut"
}

func (e *RideTimedOutEvent) Validate() error {
	if e.RideID == uuid.Nil {
		return errors.NewValidationErrorf("RideID cannot be empty")
	}

	return nil
}

var _ ports.MessageInterface = (*RideTimedOutEvent)(nil)
