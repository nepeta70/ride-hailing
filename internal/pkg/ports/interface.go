package ports

import "context"

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
