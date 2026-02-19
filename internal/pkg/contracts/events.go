package contracts

import (
	"time"
)

type EventMessage struct {
	ID        string    `json:"id"`
	EventType string    `json:"eventType"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"message"`
}

type Event interface {
	EventType() string
}
