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

func NewCacheService(client *redis.Client) ports.CacheService {
	return &redisCache{client: client}
}

func (s *redisCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, dest any, fetch func() (any, error)) error {
	// 1. Try Cache
	val, err := s.client.Get(ctx, key).Result()
	if err == nil {
		return json.Unmarshal([]byte(val), dest)
	}

	// 2. Cache Miss - Fetch from source
	data, err := fetch()
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
