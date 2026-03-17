// internal/location/service/service.go
package service

import (
	"context"

	"github.com/google/uuid"
	common "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type LocationServiceOpts struct {
	Config    *config.Config
	Repo      ports.LocationRepository
	Telemetry pkgPorts.TelemetryProvider
}

func (opts *LocationServiceOpts) Validate() error {
	if opts.Config == nil {
		return errors.NewValidationErrorf("config cannot be nil")
	}
	if opts.Repo == nil {
		return errors.NewValidationErrorf("repo cannot be nil")
	}
	if opts.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry cannot be nil")
	}

	return nil
}

type LocationService struct {
	config            *config.Config
	repo              ports.LocationRepository
	retryFactory      pkgPorts.RetrierFactoryInterface
	telemetryProvider pkgPorts.TelemetryProvider
}

func NewLocationService(opts *LocationServiceOpts) *LocationService {
	if err := opts.Validate(); err != nil {
		panic(err)
	}
	return &LocationService{
		config:            opts.Config,
		repo:              opts.Repo,
		telemetryProvider: opts.Telemetry,
	}
}

func (s *LocationService) Update(ctx context.Context, req *UpdateDriverLocationRequest) error {
	ctx, span := s.telemetryProvider.Tracer().Start(ctx, "UpdateDriverLocation",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("driver.id", req.DriverID.String()),
			attribute.String("sender.type", req.SenderType.String()),
			attribute.Float64("location.latitude", req.Coordinates.Latitude),
			attribute.Float64("location.longitude", req.Coordinates.Longitude),
		),
	)
	defer span.End()
	if err := ctx.Err(); err != nil {
		span.RecordError(err)
		return errors.ErrContextError
	}
	if err := req.Validate(); err != nil {
		span.RecordError(err)
		return err // Returns the validation Error
	}

	// 2. Map to Domain (The Geohash-only object you wanted)
	loc := &domain.DriverLocation{
		UserID:     req.DriverID,
		SenderType: req.SenderType,
		Coordinates: common.Coordinates{
			Latitude:  req.Coordinates.Latitude,
			Longitude: req.Coordinates.Longitude,
		},
		Accuracy:   req.Accuracy,
		Heading:    req.Heading,
		Speed:      req.Speed,
		Status:     req.Status,
		CapturedAt: req.CapturedAt,
	}

	return s.repo.SaveDriverLocation(ctx, loc)
}

func (s *LocationService) Get(ctx context.Context, userID uuid.UUID) (*domain.DriverLocation, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}
	return s.repo.GetDriverLocationAndStatus(ctx, userID)
}

func (s *LocationService) RemoveUserLocation(ctx context.Context, userID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}
	return s.repo.RemoveUserLocation(ctx, userID)
}

func (s *LocationService) SearchNearby(ctx context.Context, coordinates *common.Coordinates) ([]*domain.DriverLocation, error) {
	ctx, span := s.telemetryProvider.Tracer().Start(ctx, "UpdateDriverLocation",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.Float64("search.latitude", coordinates.Latitude),
			attribute.Float64("search.longitude", coordinates.Longitude),
		),
	)
	defer span.End()
	if err := ctx.Err(); err != nil {
		span.RecordError(err)
		return nil, errors.ErrContextError
	}

	radius := s.config.Logic.MinRadiusSearchKm

	var drivers []*domain.DriverLocation
	var err error
	for attempt := 1; radius <= s.config.Logic.MaxRadiusSearchKm; attempt++ {
		span.AddEvent("Searching for nearby drivers", trace.WithAttributes(
			attribute.Int("attempt", attempt),
			attribute.Float64("radius_km", float64(radius)),
		))
		s.telemetryProvider.Logger().DebugContext(ctx, "Searching for nearby drivers", "radius", radius, "attempt", attempt)
		drivers, err = s.repo.SearchNearby(ctx, coordinates, radius)
		if err == nil {
			if len(drivers) > 0 {
				return drivers, nil
			}
			s.telemetryProvider.Logger().DebugContext(ctx, "No drivers found within radius, expanding search", "radius", radius)
		}
		radius *= 2 // Exponential backoff for radius
	}

	if len(drivers) == 0 {
		return nil, errors.NewErrNotFoundf("no drivers found within radius %.2f km", radius)
	}
	return drivers, err
}

func (s *LocationService) UpdateDriverStatus(ctx context.Context, req *UpdateDriverStatusRequest) error {
	ctx, span := s.telemetryProvider.Tracer().Start(ctx, "UpdateDriverStatus",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("driver.id", req.DriverID.String()),
			attribute.String("status", req.Status.String()),
		),
	)
	defer span.End()

	if err := ctx.Err(); err != nil {
		span.RecordError(err)
		return errors.ErrContextError
	}
	if err := req.Validate(); err != nil {
		span.RecordError(err)
		return err
	}

	status := &domain.DriverStatusUpdate{
		DriverID:        req.DriverID,
		Status:          req.Status,
		StatusUpdatedAt: req.StatusUpdatedAt,
	}

	return s.repo.SaveDriverStatus(ctx, status)
}
