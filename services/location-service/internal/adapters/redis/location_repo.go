package redisStore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

const (
	locationIndexKeyPrefix = "locations:index"
	userLocationKeyPrefix  = "locations:user-location:"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func userLocationKey(userID string) string {
	return userLocationKeyPrefix + userID
}

// Save implements the ports.LocationRepository interface
func (r *RedisRepository) Save(ctx context.Context, loc *domain.UserLocation) error {
	// Specific metadata for this entity (HASH)
	userLocationKey := userLocationKey(loc.UserID)

	data, err := json.Marshal(loc)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	pipe := r.client.Pipeline()

	if loc.UserType == domain.UserTypeDriver {
		// 1. Add to Geospatial Index for Drivers
		pipe.GeoAdd(ctx, locationIndexKeyPrefix, &redis.GeoLocation{
			Longitude: loc.Coordinates.Longitude,
			Latitude:  loc.Coordinates.Latitude,
			Name:      loc.UserID,
		})
	}

	// 3. Save Metadata in a Hash
	pipe.HSet(ctx, userLocationKey, "data", data)

	// Set TTL so we don't leak memory for offline drivers (e.g., 1 hour)
	pipe.Expire(ctx, userLocationKey, time.Hour)

	_, err = pipe.Exec(ctx)
	return errors.NewTransientErrorf("tx pipelined fleet swap failed: %w", err)
}

func (r *RedisRepository) Get(ctx context.Context, userID string) (*domain.UserLocation, error) {
	key := userLocationKey(userID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, errors.NewTransientErrorf("Redis error: %w", err)
	}

	var loc domain.UserLocation
	if err := json.Unmarshal(data, &loc); err != nil {
		return nil, errors.NewErrJSONUnmarshal(err)
	}

	return &loc, nil
}

func (r *RedisRepository) RemoveUserLocation(ctx context.Context, userID string) error {
	// Remove metadata
	userLocationKey := userLocationKey(userID)
	if err := r.client.Del(ctx, userLocationKey).Err(); err != nil {
		return errors.NewTransientErrorf("failed to delete location metadata: %w", err)
	}
	// Note: In a real implementation, you'd also want to remove the entry from the ZSET.
	return nil
}
