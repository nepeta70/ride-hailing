package contracts

import (
	"time"
)

type EventType string

type Event interface {
	EventType() EventType
}

type EventMessage struct {
	ID        string    `json:"id"`
	EventType EventType `json:"eventType"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"message"`
}
