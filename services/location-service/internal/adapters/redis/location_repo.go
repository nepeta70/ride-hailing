package redisStore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/internal/pkg/logging"
	redisClient "github.com/nepeta70/ride-hailing/internal/pkg/redis"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

const (
	locationIndexKeyPrefix = "locations:index"
	userLocationKeyPrefix  = "locations:user-location:"
)

type RedisRepository struct {
	client *redis.Client
	cfg    *config.Config
	logger logging.Logger
}

func NewRedisRepository(cfg *config.Config, client *redisClient.RedisClient, logger logging.Logger) *RedisRepository {
	return &RedisRepository{client: client.Rdb, cfg: cfg, logger: logger}
}

func userLocationKey(userID string) string {
	return userLocationKeyPrefix + userID
}

// Save implements the ports.LocationRepository interface
func (r *RedisRepository) Save(ctx context.Context, loc *domain.UserLocation) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.StandardTimeout)
	defer cancel()
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

	// Set TTL so we don't leak memory for offline users
	pipe.Expire(ctx, userLocationKey, time.Duration(r.cfg.Logic.LocationTTLSeconds)*time.Second)

	if _, err = pipe.Exec(ctx); err != nil {
		return errors.NewTransientErrorf("tx pipelined fleet swap failed: %w", err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, userID string) (*domain.UserLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.StandardTimeout)
	defer cancel()
	key := userLocationKey(userID)

	data, err := r.client.HGet(ctx, key, "data").Bytes()
	if err == redis.Nil {
		// Not found, return nil, nil
		return nil, errors.NewErrNotFound("location data for userID " + userID)
	}
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
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.StandardTimeout)
	defer cancel()
	// Remove metadata
	userLocationKey := userLocationKey(userID)
	if err := r.client.Del(ctx, userLocationKey).Err(); err != nil {
		return errors.NewTransientErrorf("failed to delete location metadata: %w", err)
	}

	// Remove from geospatial index (ZSET)
	if err := r.client.ZRem(ctx, locationIndexKeyPrefix, userID).Err(); err != nil {
		return errors.NewTransientErrorf("failed to remove from geospatial index: %w", err)
	}

	return nil
}

func (r *RedisRepository) SearchNearby(ctx context.Context, coordinates domain.Coordinates, radiusKm float32) ([]*domain.UserLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.StandardTimeout)
	defer cancel()
	// 1. Query Geospatial Index
	geoResults, err := r.client.GeoRadius(ctx, locationIndexKeyPrefix, coordinates.Longitude, coordinates.Latitude, &redis.GeoRadiusQuery{
		Radius:    float64(radiusKm),
		Unit:      "km",
		WithCoord: false,
		WithDist:  false,
		Count:     r.cfg.Logic.TopKNearby,
		Sort:      "ASC",
	}).Result()
	if err != nil {
		return nil, errors.NewTransientErrorf("redis georadius query failed: %w", err)
	}
	var results []*domain.UserLocation
	for _, geoLocation := range geoResults {
		loc, err := r.Get(ctx, geoLocation.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				r.asyncRemoveFromIndex(geoLocation.Name)
				continue // Skip missing entries
			}
			return nil, errors.NewTransientErrorf("get user location failed: %w", err)
		}
		results = append(results, loc)
	}
	return results, nil
}

func (r *RedisRepository) asyncRemoveFromIndex(userID string) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), r.cfg.Timeouts.StandardTimeout)
		defer cancel()
		r.client.ZRem(bgCtx, locationIndexKeyPrefix, userID)
	}()
}
