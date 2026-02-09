package ports

import (
	"context"
	"time"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type HealthProvider interface {
	HealthCheck(ctx context.Context) error
	ServiceName() string
}

type CacheService interface {
	// GetOrSet handles the lookup, miss, and backfill logic
	GetOrSet(ctx context.Context, key string, ttl time.Duration, dest any, fetch func() (any, error)) error
}
