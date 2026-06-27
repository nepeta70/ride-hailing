package rdstore

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/redis/go-redis/v9"
)

const redisServiceName = "RedisClient"

type RedisClient struct {
	Rdb       *redis.Client
	config    *RedisConfig
	retrier   ports.RetrierInterface
	telemetry ports.TelemetryProvider
}

// NewClient returns our wrapped client
func NewClient(cfg *RedisConfig, retrierFactory ports.RetrierFactoryInterface, telemetry ports.TelemetryProvider) (*RedisClient, error) {
	ctx, span := telemetry.Tracer().Start(context.Background(), "Redis:Initialize",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.Bool("service.init", true)),
	)
	defer span.End()

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		DialTimeout:  cfg.DialTimeout,
		PoolFIFO:     true,
		PoolSize:     cfg.PoolSize,
		PoolTimeout:  cfg.PoolTimeout,
		MinIdleConns: cfg.MinIdle,
	})

	// Initial check
	ctx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()

	retrier := retrierFactory.NewExponentialBackoffRetrier(redisServiceName, cfg.DialTimeout)
	err := retrier.Do(ctx, func(ctx context.Context) error {
		err := rdb.Ping(ctx).Err()
		if err != nil {
			telemetry.Metrics().DependencyFailure(redisServiceName, "initial_check", "error")
			telemetry.Logger().Error("Redis initial check failed", "error", err)
			return errors.NewTransientErrorf("redis not ready: %w", err)
		}
		return nil
	})

	if err != nil {
		telemetry.Metrics().DependencyFailure(redisServiceName, "initial_check", "error")
		telemetry.Logger().Error("Redis initialization exhausted", "error", err)
		return nil, errors.NewPermanentErrorf("redis initialization exhausted: %w", err)
	}

	return &RedisClient{Rdb: rdb, config: cfg, telemetry: telemetry, retrier: retrier}, nil
}

func (c *RedisClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.DialTimeout)
	defer cancel()

	if err := c.Rdb.Ping(ctx).Err(); err != nil {
		c.telemetry.Logger().Error("Redis healthcheck failed", "error", err)
		c.telemetry.Metrics().DependencyFailure(c.ServiceName(), "health_check", "error")
		return errors.NewTransientErrorf("redis healthcheck failed: %w", err)
	}
	return nil
}

func (c *RedisClient) Close() error {
	return c.Rdb.Close()
}

func (c *RedisClient) ServiceName() string {
	return redisServiceName
}

func (c *RedisClient) TraceSpan(ctx context.Context, spanName string, operation string, key string) (context.Context, trace.Span) {
	tracer := c.telemetry.Tracer()
	ctx, span := tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", operation),
			attribute.String("db.key", key),
		),
	)
	return ctx, span
}

var _ ports.HealthProvider = (*RedisClient)(nil)
