package rdstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
	"github.com/redis/go-redis/v9"
)

const (
	locationIndexKeyPrefix = "locations:index"
	userLocationKeyPrefix  = "locations:user-location:"
)

type RedisRepository struct {
	client *redis.Client
	cfg    *config.Config
	logger pkgPorts.Logger
}

func NewRedisRepository(cfg *config.Config, client *rdstore.RedisClient, logger pkgPorts.Logger) *RedisRepository {
	return &RedisRepository{client: client.Rdb, cfg: cfg, logger: logger}
}

func userLocationKey(userID uuid.UUID) string {
	return userLocationKeyPrefix + userID.String()
}

// Save implements the ports.LocationRepository interface
func (r *RedisRepository) Save(ctx context.Context, loc *domain.UserLocation) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	// Specific metadata for this entity (HASH)
	userLocationKey := userLocationKey(loc.UserID)

	data, err := json.Marshal(loc)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	pipe := r.client.Pipeline()

	if loc.UserType == enums.UserTypeDriver {
		// 1. Add to Geospatial Index for Drivers
		pipe.GeoAdd(ctx, locationIndexKeyPrefix, &redis.GeoLocation{
			Longitude: loc.Coordinates.Longitude,
			Latitude:  loc.Coordinates.Latitude,
			Name:      loc.UserID.String(),
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

func (r *RedisRepository) Get(ctx context.Context, userID uuid.UUID) (*domain.UserLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	key := userLocationKey(userID)

	data, err := r.client.HGet(ctx, key, "data").Bytes()
	if err == redis.Nil {
		// Not found, return nil, nil
		return nil, errors.NewErrNotFound("location data for userID " + userID.String())
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

func (r *RedisRepository) RemoveUserLocation(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
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
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	// 1. Query Geospatial Index
	query := &redis.GeoSearchQuery{
		Radius:     float64(radiusKm),
		Longitude:  coordinates.Longitude,
		Latitude:   coordinates.Latitude,
		RadiusUnit: "km",
		Count:      r.cfg.Logic.TopKNearby,
		Sort:       "ASC",
	}
	geoResults, err := r.client.GeoSearch(ctx, locationIndexKeyPrefix, query).Result()

	if err != nil {
		return nil, errors.NewTransientErrorf("redis georadius query failed: %w", err)
	}
	var results []*domain.UserLocation
	for _, geoLocation := range geoResults {
		loc, err := r.Get(ctx, uuid.MustParse(geoLocation))
		if err != nil {
			if errors.IsNotFound(err) {
				r.asyncRemoveFromIndex(geoLocation)
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
		bgCtx, cancel := context.WithTimeout(context.Background(), r.cfg.Timeouts.RequestTimeout)
		defer cancel()
		r.client.ZRem(bgCtx, locationIndexKeyPrefix, userID)
	}()
}

func (r *RedisRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		return errors.NewTransientErrorf("redis ping failed: %w", err)
	}
	return nil
}

var _ ports.LocationRepository = (*RedisRepository)(nil)
