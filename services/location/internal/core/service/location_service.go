// internal/location/service/service.go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
)

type LocationServiceOpts struct {
	Config  *config.Config
	Repo    ports.LocationRepository
	Logger  pkgPorts.Logger
	Metrics pkgPorts.Metrics
}

func (opts *LocationServiceOpts) Validate() error {
	if opts.Config == nil {
		return errors.NewValidationErrorf("config cannot be nil")
	}
	if opts.Repo == nil {
		return errors.NewValidationErrorf("repo cannot be nil")
	}
	if opts.Logger == nil {
		return errors.NewValidationErrorf("logger cannot be nil")
	}
	if opts.Metrics == nil {
		return errors.NewValidationErrorf("metrics cannot be nil")
	}
	return nil
}

type LocationService struct {
	config       *config.Config
	repo         ports.LocationRepository
	retryFactory pkgPorts.RetrierFactoryInterface
	logger       pkgPorts.Logger
	metrics      pkgPorts.Metrics
}

func NewLocationService(opts *LocationServiceOpts) *LocationService {
	if err := opts.Validate(); err != nil {
		panic(err)
	}
	return &LocationService{
		config:  opts.Config,
		repo:    opts.Repo,
		logger:  opts.Logger,
		metrics: opts.Metrics,
	}
}

func (s *LocationService) Update(ctx context.Context, req *UpdateRequest) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}
	if err := req.Validate(); err != nil {
		return err // Returns the Business Error
	}

	// 2. Map to Domain (The Geohash-only object you wanted)
	loc := &domain.DriverLocation{
		UserID:   req.UserID,
		UserType: req.UserType,
		Coordinates: domain.Coordinates{
			Latitude:  req.Coordinates.Latitude,
			Longitude: req.Coordinates.Longitude,
		},
		Accuracy:   req.Accuracy,
		Heading:    req.Heading,
		Speed:      req.Speed,
		Status:     req.Status,
		CapturedAt: req.CapturedAt,
	}

	return s.repo.Save(ctx, loc)
}

func (s *LocationService) Get(ctx context.Context, userID uuid.UUID) (*domain.DriverLocation, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}
	return s.repo.Get(ctx, userID)
}

func (s *LocationService) RemoveUserLocation(ctx context.Context, userID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}
	return s.repo.RemoveUserLocation(ctx, userID)
}

func (s *LocationService) SearchNearby(ctx context.Context, req *SearchNearbyRequest) ([]*domain.DriverLocation, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	radius := s.config.Logic.MinRadiusSearchKm

	var drivers []*domain.DriverLocation
	var err error
	for attempt := 1; radius <= s.config.Logic.MaxRadiusSearchKm; attempt++ {
		drivers, err = s.repo.SearchNearby(ctx, req.Coordinates, radius)
		if err == nil {
			if len(drivers) > 0 {
				return drivers, nil
			}
			return nil, errors.NewTransientErrorf("no drivers found within radius %.2f km", radius)
		}
		radius *= 2 // Exponential backoff for radius
		if radius > s.config.Logic.MaxRadiusSearchKm {
			radius = s.config.Logic.MaxRadiusSearchKm
		}
	}

	return drivers, err
}
