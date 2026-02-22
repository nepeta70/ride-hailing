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
	Payload   any               `json:"message"`
	Headers   map[string][]byte `json:"-"`
}

func NewEventMessage(event Event) *EventMessage {
	return &EventMessage{
		EventType: event.EventType(),
		Payload:   event,
		Headers:   map[string][]byte{"event-type": []byte(event.EventType())},
	}
}

func (e *EventMessage) AddHeader(key string, value string) {
	if e.Headers == nil {
		e.Headers = make(map[string][]byte)
	}
	e.Headers[key] = []byte(value)
}

func (e *EventMessage) AddHeaders(headers map[string][]byte) {
	if e.Headers == nil {
		e.Headers = headers
	} else {
		maps.Copy(e.Headers, headers)
	}
}
