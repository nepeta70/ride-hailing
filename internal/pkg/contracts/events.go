package contracts

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
)

type EventMessage struct {
	ID            string              `json:"id"`
	AggregateType enums.AggregateType `json:"aggregateType"`
	AggregateID   string              `json:"aggregateId"`
	Version       int                 `json:"version"`
	EventType     string              `json:"eventType"`
	Timestamp     time.Time           `json:"timestamp"`
	Payload       any                 `json:"message"`
}
