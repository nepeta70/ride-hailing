package rdstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/redis/go-redis/v9"
)

type RedisDb[T any] struct {
	client *redis.Client
	cfg    *config.BaseConfig
	logger ports.Logger
}

func NewRedisStorage[T any](cfg *config.BaseConfig, client *RedisClient, logger ports.Logger) *RedisDb[T] {
	return &RedisDb[T]{
		cfg:    cfg,
		client: client.Rdb,
		logger: logger,
	}
}

func (r *RedisDb[T]) Get(ctx context.Context, key string) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.NewErrNotFound("Fare not found")
		}
		return nil, err
	}

	var value T
	err = json.Unmarshal([]byte(data), &value)
	if err != nil {
		return nil, errors.NewErrJSONUnmarshal(err)
	}

	return &value, nil
}

func (r *RedisDb[T]) Save(ctx context.Context, key string, value *T, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()

	data, err := json.Marshal(value)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}
