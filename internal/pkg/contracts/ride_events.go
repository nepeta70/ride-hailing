package contracts

import (
	"github.com/google/uuid"
	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
)

const (
	EventTypeRideRequested EventType = "RideRequested"
	EventTypeRideAccepted  EventType = "RideAccepted"
	EventTypeRideRejected  EventType = "RideRejected"
	EventTypeRideCanceled  EventType = "RideCanceled"
	EventTypeRideCompleted EventType = "RideCompleted"
	EventTypeRideStarted   EventType = "RideStarted"
	EventTypeRideTimedOut  EventType = "RideTimedOut"
)

type RideAcceptedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideAcceptedEvent) EventType() EventType {
	return EventTypeRideAccepted
}

var _ Event = (*RideAcceptedEvent)(nil)

type RideCanceledEvent struct {
	RequestID uuid.UUID
	RiderID   uuid.UUID
	RideID    uuid.UUID
}

func (e *RideCanceledEvent) EventType() EventType {
	return EventTypeRideCanceled
}

var _ Event = (*RideCanceledEvent)(nil)

type RideCompletedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideCompletedEvent) EventType() EventType {
	return EventTypeRideCompleted
}

var _ Event = (*RideCompletedEvent)(nil)

type RideRejectedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideRejectedEvent) EventType() EventType {
	return EventTypeRideRejected
}

var _ Event = (*RideRejectedEvent)(nil)

type RideRequestedEvent struct {
	RideID          uuid.UUID
	RequestID       uuid.UUID
	RiderID         uuid.UUID
	PickupLocation  *core.Coordinates
	DropoffLocation *core.Coordinates

	ServiceType string
	Fare        float64
	Currency    string
}

func (e *RideRequestedEvent) EventType() EventType {
	return EventTypeRideRequested
}

var _ Event = (*RideRequestedEvent)(nil)

type RideStartedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideStartedEvent) EventType() EventType {
	return EventTypeRideStarted
}

var _ Event = (*RideStartedEvent)(nil)

type RideTimedOutEvent struct {
	RequestID uuid.UUID
	RideID    uuid.UUID
}

func (e *RideTimedOutEvent) EventType() EventType {
	return EventTypeRideTimedOut
}

var _ Event = (*RideTimedOutEvent)(nil)
