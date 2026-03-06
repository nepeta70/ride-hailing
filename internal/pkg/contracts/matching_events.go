package contracts

import "github.com/google/uuid"

const (
	EventTypeRideMatched     EventType = "RideMatched"
	EventTypeMatchingTimeout EventType = "MatchingTimeout"
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

type MatchingTimeoutEvent struct {
	RideID uuid.UUID
}

func (e *MatchingTimeoutEvent) EventType() EventType {
	return EventTypeMatchingTimeout
}

var _ Event = (*MatchingTimeoutEvent)(nil)
