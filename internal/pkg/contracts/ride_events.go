package contracts

import (
	"github.com/google/uuid"
)

type RideAcceptedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideAcceptedEvent) EventType() string {
	return "RideAccepted"
}

var _ Event = (*RideAcceptedEvent)(nil)

type RideCanceledEvent struct {
	RequestID uuid.UUID
	RiderID   uuid.UUID
	RideID    uuid.UUID
}

func (e *RideCanceledEvent) EventType() string {
	return "RideCanceled"
}

var _ Event = (*RideCanceledEvent)(nil)

type RideCompletedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideCompletedEvent) EventType() string {
	return "RideCompleted"
}

var _ Event = (*RideCompletedEvent)(nil)

type RideRejectedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideRejectedEvent) EventType() string {
	return "RideRejected"
}

var _ Event = (*RideRejectedEvent)(nil)

type RideRequestedEvent struct {
	RequestID       uuid.UUID
	RiderID         uuid.UUID
	PickupLocation  string
	DropoffLocation string
	ServiceType     string
	Fare            float64
	Currency        string
}

func (e *RideRequestedEvent) EventType() string {
	return "RideRequested"
}

var _ Event = (*RideRequestedEvent)(nil)

type RideStartedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideStartedEvent) EventType() string {
	return "RideStarted"
}

var _ Event = (*RideStartedEvent)(nil)
