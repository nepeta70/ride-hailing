package contracts

import "maps"

type EventType string

func (e EventType) String() string {
	return string(e)
}

type Event interface {
	EventType() EventType
}

type EventMessage struct {
	EventType EventType         `json:"event-type"`
	Payload   Event             `json:"message"`
	Headers   map[string]string `json:"-"`
}

func NewEventMessage(event Event) *EventMessage {
	return &EventMessage{
		EventType: event.EventType(),
		Payload:   event,
		Headers:   map[string]string{"event-type": event.EventType().String()},
	}
}

func (e *EventMessage) AddHeader(key string, value string) {
	if e.Headers == nil {
		e.Headers = make(map[string]string, 10)
	}
	e.Headers[key] = value
}

func (e *EventMessage) AddHeaders(headers map[string]string) {
	if e.Headers == nil {
		e.Headers = headers
	} else {
		maps.Copy(e.Headers, headers)
	}
}
