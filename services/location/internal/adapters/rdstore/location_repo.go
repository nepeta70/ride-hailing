package rdstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
	"github.com/redis/go-redis/v9"
)

const (
	indexKey                = "locations:index"
	driverLocationKeyPrefix = "locations:driver:"
)

type LocationRepository struct {
	client  *redis.Client
	cfg     *config.Config
	logger  pkgPorts.Logger
	metrics pkgPorts.Metrics
}

func NewLocationRepository(cfg *config.Config, client *rdstore.RedisClient, logger pkgPorts.Logger, metrics pkgPorts.Metrics) *LocationRepository {
	return &LocationRepository{client: client.Rdb, cfg: cfg, logger: logger, metrics: metrics}
}

func driverLocationKey(userID uuid.UUID) string {
	return driverLocationKeyPrefix + userID.String()
}

// Save implements the ports.LocationRepository interface
func (r *LocationRepository) Save(ctx context.Context, loc *domain.DriverLocation) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	// Specific metadata for this entity (HASH)
	driverLocationKey := driverLocationKey(loc.UserID)

	data, err := json.Marshal(loc)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	userID := loc.UserID.String()

	pipe := r.client.Pipeline()

	// Save Metadata in a Hash
	pipe.HSet(ctx, driverLocationKey, "data", data, "status", string(loc.Status))

	// Add to Geospatial Index for Drivers if available
	if loc.Status == contracts.DriverStatusAvailable {
		pipe.GeoAdd(ctx, indexKey, &redis.GeoLocation{
			Longitude: loc.Coordinates.Longitude,
			Latitude:  loc.Coordinates.Latitude,
			Name:      userID,
		})
	} else {
		// If not available, remove from geospatial index to prevent showing in nearby searches
		pipe.ZRem(ctx, indexKey, userID)
	}

	// Set TTL so we don't leak memory for offline users
	pipe.Expire(ctx, driverLocationKey, time.Duration(r.cfg.Logic.LocationTTLSeconds)*time.Second)

	if _, err = pipe.Exec(ctx); err != nil {
		r.metrics.DependencyFailure("redis", "pipe_exec", "save_location")
		return errors.NewTransientErrorf("tx pipelined fleet swap failed: %w", err)
	}
	return nil
}

func (r *LocationRepository) Get(ctx context.Context, userID uuid.UUID) (*domain.DriverLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	key := driverLocationKey(userID)

	data, err := r.client.HGet(ctx, key, "data").Bytes()
	if err == redis.Nil {
		// Not found, return nil, nil
		return nil, errors.NewErrNotFound("location data for userID " + userID.String())
	}
	if err != nil {
		r.metrics.DependencyFailure("redis", "hget", "get_location")
		return nil, errors.NewTransientErrorf("Redis error: %w", err)
	}

	var loc domain.DriverLocation
	if err := json.Unmarshal(data, &loc); err != nil {
		return nil, errors.NewErrJSONUnmarshal(err)
	}

	return &loc, nil
}

func (r *LocationRepository) RemoveUserLocation(ctx context.Context, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	// Remove metadata
	if err := r.client.Del(ctx, driverLocationKey(userID)).Err(); err != nil {
		r.metrics.DependencyFailure("redis", "del", "remove_user_location")
		return errors.NewTransientErrorf("failed to delete location metadata: %w", err)
	}

	// Remove from geospatial index (ZSET)
	if err := r.client.ZRem(ctx, indexKey, userID.String()).Err(); err != nil {
		r.metrics.DependencyFailure("redis", "zrem", "remove_user_location")
		return errors.NewTransientErrorf("failed to remove from geospatial index: %w", err)
	}

	return nil
}

func (r *LocationRepository) SearchNearby(ctx context.Context, coordinates *core.Coordinates, radiusKm float32) ([]*domain.DriverLocation, error) {
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
	geoResults, err := r.client.GeoSearch(ctx, indexKey, query).Result()

	if err != nil {
		r.metrics.DependencyFailure("redis", "geosearch", "search_nearby")
		return nil, errors.NewTransientErrorf("redis georadius query failed: %w", err)
	}

	if len(geoResults) == 0 {
		return []*domain.DriverLocation{}, nil
	}

	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.SliceCmd)
	for _, driverID := range geoResults {
		cmds[driverID] = pipe.HMGet(ctx, driverLocationKeyPrefix+driverID, "status", "data")
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		r.metrics.DependencyFailure("redis", "pipeline", "search_nearby")
		return nil, errors.NewTransientErrorf("pipeline execution failed: %w", err)
	}

	var results []*domain.DriverLocation
	for locationKey, cmd := range cmds {
		parts, err := cmd.Result()
		if err != nil || len(parts) < 2 || parts[0] == nil {
			continue
		}

		status := parts[0].(string)

		if !contracts.DriverStatusAvailable.Equals(status) {
			r.asyncRemoveFromIndex(locationKey)
			continue
		}
		var loc domain.DriverLocation
		if err := json.Unmarshal([]byte(parts[1].(string)), &loc); err != nil {
			continue
		}

		results = append(results, &loc)
	}
	return results, nil
}

func (r *LocationRepository) asyncRemoveFromIndex(userID string) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), r.cfg.Timeouts.RequestTimeout)
		defer cancel()
		if err := r.client.ZRem(bgCtx, indexKey, userID).Err(); err != nil {
			r.metrics.DependencyFailure("redis", "zrem", "async_remove_from_index")
		}
	}()
}

func (r *LocationRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		r.metrics.DependencyFailure("redis", "ping", "health_check")
		return errors.NewTransientErrorf("redis ping failed: %w", err)
	}
	return nil
}

var _ ports.LocationRepository = (*LocationRepository)(nil)
