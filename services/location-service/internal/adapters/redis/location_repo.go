package redisStore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

const (
	locationIndexKeyPrefix = "locations:index"
	locationMetaKeyPrefix  = "locations:meta:"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func locationMetaKey(entityID string) string {
	return fmt.Sprintf("%s%s", locationMetaKeyPrefix, entityID)
}

// Save implements the ports.LocationRepository interface
func (r *RedisRepository) Save(ctx context.Context, loc *domain.Location) error {

	// Specific metadata for this entity (HASH)
	metaKey := locationMetaKey(loc.EntityID)

	data, err := json.Marshal(loc)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	pipe := r.client.Pipeline()

	// 2. Add to Sorted Set for Geohash indexing
	// Redis ZSETs allow us to store the Geohash as the 'member'
	// We use the timestamp as the score so we can clean up old data easily
	pipe.ZAdd(ctx, locationIndexKeyPrefix, redis.Z{
		Score:  float64(loc.CapturedAt.Unix()),
		Member: fmt.Sprintf("%s:%s", loc.Geohash, loc.EntityID),
	})

	// 3. Save Metadata in a Hash
	pipe.HSet(ctx, metaKey, data)

	// Set TTL so we don't leak memory for offline drivers (e.g., 1 hour)
	pipe.Expire(ctx, metaKey, time.Hour)

	_, err = pipe.Exec(ctx)
	return errors.NewTransientErrorf("tx pipelined fleet swap failed: %w", err)
}

func (r *RedisRepository) Get(ctx context.Context, entityID string) (*domain.Location, error) {
	key := locationMetaKey(entityID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, errors.NewTransientErrorf("Redis error: %w", err)
	}

	var loc domain.Location
	if err := json.Unmarshal(data, &loc); err != nil {
		return nil, errors.NewErrJSONUnmarshal(err)
	}

	return &loc, nil
}
