package rdstore

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	indexKey                = "locations:index"
	driverLocationKeyPrefix = "locations:driver:"
)

type LocationRepository struct {
	client    *rdstore.RedisClient
	cfg       *config.Config
	ctxMgr    *ctxmgr.ContextManager
	telemetry pkgPorts.TelemetryProvider
}

func NewLocationRepository(cfg *config.Config, client *rdstore.RedisClient, ctxMgr *ctxmgr.ContextManager, telemetry pkgPorts.TelemetryProvider) *LocationRepository {
	return &LocationRepository{client: client, cfg: cfg, ctxMgr: ctxMgr, telemetry: telemetry}
}

func driverLocationKey(userID uuid.UUID) string {
	return driverLocationKeyPrefix + userID.String()
}

// SaveLocation implements the ports.LocationRepository interface
func (r *LocationRepository) SaveDriverLocation(ctx context.Context, loc *domain.DriverLocation) error {
	// Specific metadata for this entity (HASH)
	driverLocationKey := driverLocationKey(loc.UserID)
	ctx, span := r.client.TraceSpan(ctx, "SaveDriverLocation", "HSET", driverLocationKey)
	span.SetAttributes(
		attribute.String("driver.id", loc.UserID.String()),
		attribute.Float64("latitude", loc.Coordinates.Latitude),
		attribute.Float64("longitude", loc.Coordinates.Longitude),
	)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()

	data, err := json.Marshal(loc)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	userID := loc.UserID.String()

	pipe := r.client.Rdb.Pipeline()

	statusCmd := pipe.HGet(ctx, driverLocationKey, "status")

	pipe.HSet(ctx, driverLocationKey, "data", data)
	pipe.HSetNX(ctx, driverLocationKey, "status", contracts.DriverStatusAvailable.String()) // Default to Available if not set
	pipe.Expire(ctx, driverLocationKey, r.cfg.Logic.LocationTTL)

	if _, err = pipe.Exec(ctx); err != nil {
		span.RecordError(err)
		r.telemetry.Metrics().DependencyFailure("redis", "pipe_exec", "save_location")
		return errors.NewTransientErrorf("tx pipelined fleet swap failed: %w", err)
	}

	currStatus := statusCmd.Val()
	if currStatus == "" {
		currStatus = contracts.DriverStatusAvailable.String()
	}

	if currStatus == contracts.DriverStatusAvailable.String() {
		span.AddEvent("Adding driver to geospatial index")
		return r.client.Rdb.GeoAdd(ctx, indexKey, &redis.GeoLocation{
			Longitude: loc.Coordinates.Longitude,
			Latitude:  loc.Coordinates.Latitude,
			Name:      userID,
		}).Err()
	} else {
		span.AddEvent("Driver not available, removing from geospatial index if exists")
		return r.client.Rdb.ZRem(ctx, indexKey, userID).Err()
	}
}

func (r *LocationRepository) GetDriverLocationAndStatus(ctx context.Context, userID uuid.UUID) (*domain.DriverLocation, error) {
	key := driverLocationKey(userID)
	ctx, span := r.client.TraceSpan(ctx, "GetDriverLocationAndStatus", "HGET", key)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()

	data, err := r.client.Rdb.HGet(ctx, key, "data").Bytes()
	if err == redis.Nil {
		span.AddEvent("Driver location not found")
		// Not found, return nil, nil
		return nil, errors.NewErrNotFoundf("location data for userID %s not found", userID.String())
	}
	if err != nil {
		span.RecordError(err)
		r.telemetry.Metrics().DependencyFailure("redis", "hget", "get_location")
		return nil, errors.NewTransientErrorf("Redis error: %w", err)
	}

	var loc domain.DriverLocation
	if err := json.Unmarshal(data, &loc); err != nil {
		span.RecordError(err)
		return nil, errors.NewErrJSONUnmarshal(err)
	}

	return &loc, nil
}

func (r *LocationRepository) RemoveUserLocation(ctx context.Context, userID uuid.UUID) error {
	key := driverLocationKey(userID)
	ctx, span := r.client.TraceSpan(ctx, "RemoveUserLocation", "DEL", key)
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	// Remove metadata
	if err := r.client.Rdb.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		r.telemetry.Metrics().DependencyFailure("redis", "del", "remove_user_location")
		return errors.NewTransientErrorf("failed to delete location metadata: %w", err)
	}

	// Remove from geospatial index (ZSET)
	if err := r.client.Rdb.ZRem(ctx, indexKey, userID.String()).Err(); err != nil {
		span.RecordError(err)
		r.telemetry.Metrics().DependencyFailure("redis", "zrem", "remove_user_location")
		return errors.NewTransientErrorf("failed to remove from geospatial index: %w", err)
	}

	return nil
}

func (r *LocationRepository) SearchNearby(ctx context.Context, coordinates *core.Coordinates, radiusKm float32) ([]*domain.DriverLocation, error) {
	ctx, span := r.client.TraceSpan(ctx, "SearchNearby", "GEORADIUS", indexKey)
	defer span.End()

	reqInfo, ok := r.ctxMgr.Extract(ctx)

	if ok {
		span.SetAttributes(
			attribute.String("request.id", reqInfo.Trace.RequestID.String()),
			attribute.String("rider.id", reqInfo.Sender.ID.String()),
		)
	}
	span.SetAttributes(
		attribute.Float64("search.lat", coordinates.Latitude),
		attribute.Float64("search.lon", coordinates.Longitude),
		attribute.Float64("search.radius_km", float64(radiusKm)),
	)

	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()
	// 1. Query Geospatial Index
	query := &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Radius:     float64(radiusKm),
			Longitude:  coordinates.Longitude,
			Latitude:   coordinates.Latitude,
			RadiusUnit: "km",
			Count:      r.cfg.Logic.TopKNearby,
			Sort:       "ASC",
		},
		WithDist:  true,
		WithCoord: true,
	}
	r.telemetry.Logger().DebugContext(ctx, "Executing GeoSearch with query", "query", query)
	geoResults, err := r.client.Rdb.GeoSearchLocation(ctx, indexKey, query).Result()

	if err != nil {
		span.RecordError(err)
		r.telemetry.Logger().ErrorContext(ctx, "GeoSearch query failed", "error", err)
		r.telemetry.Metrics().DependencyFailure("redis", "geosearch", "search_nearby")
		return nil, errors.NewTransientErrorf("redis georadius query failed: %w", err)
	}

	if len(geoResults) == 0 {
		span.AddEvent("No nearby drivers found in geospatial index")
		return []*domain.DriverLocation{}, nil
	}

	pipe := r.client.Rdb.Pipeline()
	n := len(geoResults)
	geoMap := make(map[string]*redis.GeoLocation, n)
	cmds := make(map[string]*redis.SliceCmd, n)
	for _, location := range geoResults {
		span.AddEvent("Processing GeoSearch result", trace.WithAttributes(
			attribute.String("driver.id", location.Name),
			attribute.Float64("distance.km", location.Dist),
		))
		driverID := location.Name
		geoMap[driverID] = &location
		cmds[driverID] = pipe.HMGet(ctx, driverLocationKeyPrefix+driverID, "status", "data")
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		span.RecordError(err)
		r.telemetry.Logger().ErrorContext(ctx, "Pipeline execution failed", "error", err)
		r.telemetry.Metrics().DependencyFailure("redis", "pipeline", "search_nearby")
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
			span.AddEvent("Driver not available, removing from index", trace.WithAttributes(
				attribute.String("driver.id", locationKey),
				attribute.String("status", status),
			))
			// Asynchronously remove from geospatial index to prevent showing in future searches
			r.asyncRemoveFromIndex(locationKey)
			continue
		}
		var loc domain.DriverLocation
		if err := json.Unmarshal([]byte(parts[1].(string)), &loc); err != nil {
			span.RecordError(err)
			r.telemetry.Logger().ErrorContext(ctx, "Failed to unmarshal driver location", "error", err)
			continue
		}
		loc.DistanceKm = float32(geoMap[locationKey].Dist) // Distance in kilometers

		results = append(results, &loc)
	}

	span.SetStatus(codes.Ok, "candidates found")
	return results, nil
}

func (r *LocationRepository) SaveDriverStatus(ctx context.Context, status *domain.DriverStatusUpdate) error {
	key := driverLocationKey(status.DriverID)
	ctx, span := r.client.TraceSpan(ctx, "SaveDriverStatus", "HSET", key)
	defer span.End()
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	span.SetAttributes(
		attribute.String("driver.id", status.DriverID.String()),
		attribute.String("status.updated", status.Status.String()),
	)
	defer cancel()
	// Update status field in the HASH
	if err := r.client.Rdb.HSet(ctx, key, "status", status.Status.String()).Err(); err != nil {
		span.RecordError(err)
		r.telemetry.Metrics().DependencyFailure("redis", "hset", "save_driver_status")
		return errors.NewTransientErrorf("failed to update driver status: %w", err)
	}

	span.SetStatus(codes.Ok, "Driver status updated")
	return nil
}

func (r *LocationRepository) asyncRemoveFromIndex(userID string) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), r.cfg.Timeouts.RequestTimeout)
		defer cancel()
		if err := r.client.Rdb.ZRem(bgCtx, indexKey, userID).Err(); err != nil {
			r.telemetry.Logger().ErrorContext(bgCtx, "Failed to remove from geospatial index", "error", err)
			r.telemetry.Metrics().DependencyFailure("redis", "zrem", "async_remove_from_index")
		}
	}()
}

var _ ports.LocationRepository = (*LocationRepository)(nil)
