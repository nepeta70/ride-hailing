package ports

import (
	"context"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

type HealthProvider interface {
	HealthCheck(ctx context.Context) error
	ServiceName() string
}

type CacheService interface {
	// GetOrSet handles the lookup, miss, and backfill logic
	GetOrSet(ctx context.Context, key string, ttl time.Duration, dest any, fetch func(ctx context.Context) (any, error)) error
}

type GenericCache[T any] interface {
	GetOrSet(
		ctx context.Context,
		key string,
		ttl time.Duration,
		fetch func(context.Context) (T, error),
	) (T, error)
}

type MessageHandler func(ctx context.Context, headers map[string]string, msg []byte) error

type EventPublisher interface {
	HealthProvider
	Publish(ctx context.Context, topic contracts.Topic, message *contracts.EventMessage) error
	Close() error
	TopicProvider() TopicProvider
}

// EventSubscriber defines the contract for listening to external events
type EventSubscriber interface {
	Subscribe(ctx context.Context, topic contracts.Topic, handler MessageHandler) error
	Close() error
}

type EventStore interface {
	Append(ctx context.Context, tx Transaction, event *contracts.EventMessage) error
	GetStream(ctx context.Context, streamId enums.AggregateType, ID string) ([]contracts.EventMessage, error)
}

type TopicProvider interface {
	AllTopics() []string
	GetTopicForEvent(eventType string) (contracts.Topic, error)
}
