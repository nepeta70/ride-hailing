package rdstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewCacheService(client *redis.Client) (ports.CacheService, error) {
	if client == nil {
		return nil, errors.NewValidationErrorf("redis client is required")
	}
	return &redisCache{client: client}, nil
}

func (s *redisCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, dest any, fetch func(ctx context.Context) (any, error)) error {
	// 1. Try Cache
	val, err := s.client.Get(ctx, key).Result()
	if err == nil {
		return json.Unmarshal([]byte(val), dest)
	}

	// 2. Cache Miss - Fetch from source
	data, err := fetch(ctx)
	if err != nil {
		return err
	}

	// 3. Update destination pointer so the caller has the data immediately
	bytes, _ := json.Marshal(data)
	if err := json.Unmarshal(bytes, dest); err != nil {
		return errors.NewErrJSONUnmarshal(err)
	}

	// 4. Backfill Redis
	return s.client.Set(ctx, key, bytes, ttl).Err()
}

var _ ports.CacheService = (*redisCache)(nil)

type genericCache[T any] struct {
	client *RedisClient
}

func NewGenericCache[T any](client *RedisClient) (*genericCache[T], error) {
	if client == nil {
		return nil, errors.NewValidationErrorf("redis client is required")
	}
	return &genericCache[T]{client: client}, nil
}

func (c *genericCache[T]) GetOrSet(
    ctx context.Context,
    key string,
    ttl time.Duration,
    fetch func(context.Context) (T, error),
) (T, error) {
    var zero T

    val, err := c.client.Rdb.Get(ctx, key).Result()
    if err == nil {
        var out T
        if err := json.Unmarshal([]byte(val), &out); err != nil {
            return zero, err
        }
        return out, nil
    }
    if err != redis.Nil {
        return zero, err
    }

    out, err := fetch(ctx)
    if err != nil {
        return zero, err
    }

    b, err := json.Marshal(out)
    if err == nil {
        _ = c.client.Rdb.Set(ctx, key, b, ttl).Err() // don't fail main flow on cache write
    }

    return out, nil
}

var _ ports.GenericCache[any] = (*genericCache[any])(nil)
