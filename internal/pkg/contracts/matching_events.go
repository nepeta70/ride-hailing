package contracts

import "github.com/google/uuid"

const (
	EventTypeRideMatched EventType = "RideMatched"
)

type RideMatchedEvent struct {
	RequestID uuid.UUID
	DriverID  uuid.UUID
	RideID    uuid.UUID
}

func (e *RideMatchedEvent) EventType() EventType {
	return EventTypeRideMatched
}

var _ Event = (*RideMatchedEvent)(nil)
