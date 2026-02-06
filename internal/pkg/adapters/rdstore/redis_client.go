package rdstore

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/internal/pkg/resiliency/retry"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Rdb     *redis.Client
	config  *RedisConfig
	logger  ports.Logger
	retrier *retry.Retrier
}

// NewClient returns our wrapped client
func NewClient(cfg *RedisConfig, logger ports.Logger) (*RedisClient, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	retrier := retry.NewExponentialBackoffRetrierWithTimeout(cfg.DialTimeout, logger)
	err := retrier.Do(ctx, func() error {
		err := rdb.Ping(ctx).Err()
		if err != nil {
			return errors.NewTransientErrorf("redis not ready: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, errors.NewPermanentErrorf("redis initialization exhausted: %w", err)
	}

	return &RedisClient{Rdb: rdb, config: cfg, logger: logger, retrier: retrier}, nil
}

func (c *RedisClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.DialTimeout)
	defer cancel()

	if err := c.Rdb.Ping(ctx).Err(); err != nil {
		return errors.NewTransientErrorf("redis healthcheck failed: %w", err)
	}
	return nil
}

func (c *RedisClient) Close() error {
	return c.Rdb.Close()
}

func (c *RedisClient) ServiceName() string {
	return "RedisClient"
}

var _ ports.HealthProvider = (*RedisClient)(nil)
