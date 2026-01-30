package redis

import (
	"context"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Rdb *redis.Client
}

// NewClient returns our wrapped client
func NewClient(cfg RedisConfig) (*RedisClient, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, errors.NewPermanentErrorf("redis connection failed: %w", err)
	}

	return &RedisClient{Rdb: rdb}, nil
}

func (c *RedisClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
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
