package ports

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
)

// Message is the base interface for all grain messages (commands, queries, responses)
type Message any

// Grain defines a virtual actor with lifecycle hooks
type Grain interface {
	// OnActivate is called when grain is loaded into memory
	// Receives its identity to know which entity it represents
	OnActivate(ctx context.Context, identity *grain.GrainIdentity) error

	// OnReceive handles incoming command messages
	// Returns response message and error
	OnReceive(ctx context.Context, msg Message) (Message, error)

	// OnDeactivate is called before grain is removed from memory
	OnDeactivate(ctx context.Context) error

	GetStatus() any
}

type Silo interface {
	GetOrActivate(ctx context.Context, identity *grain.GrainIdentity) (GrainRef, error)

	// RegisterFactory attaches a grain creation function to an aggregate type.
	RegisterFactory(kind enums.AggregateType, factory GrainFactory)

	// Tell sends a one-way signal to a grain (fire and forget).
	Tell(ctx context.Context, identity *grain.GrainIdentity, msg Message) error

	// Ask sends a signal to a grain and waits for the resulting state or response.
	Ask(ctx context.Context, identity *grain.GrainIdentity, msg Message) (Message, error)

	// Deactivate gracefully shuts down a specific grain and removes it from memory.
	Deactivate(ctx context.Context, identity *grain.GrainIdentity) error

	// Reset clears the entire silo, deactivating all hosted grains.
	Reset(ctx context.Context) error
}

// GrainFactory creates new grain instances with their identity injected
type GrainFactory func(identity *grain.GrainIdentity) Grain

type MessageInterface interface {
	MessageName() string
	Message
}

type GrainRef interface {
	Tell(ctx context.Context, msg Message) error
	Ask(ctx context.Context, msg Message) (Message, error)
	Identity() *grain.GrainIdentity
	GetStatus() any
}
