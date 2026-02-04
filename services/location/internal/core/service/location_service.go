// internal/location/service/service.go
package service

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location/internal/ports"
)

type LocationService struct {
	repo ports.LocationRepository
}

func NewLocationService(r ports.LocationRepository) *LocationService {
	return &LocationService{repo: r}
}

func (s *LocationService) Update(ctx context.Context, req *UpdateRequest) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}
	if err := req.Validate(); err != nil {
		return err // Returns the Business Error
	}

	// 2. Map to Domain (The Geohash-only object you wanted)
	loc := &domain.UserLocation{
		UserID:   req.UserID,
		UserType: req.UserType,
		Coordinates: domain.Coordinates{
			Latitude:  req.Coordinates.Latitude,
			Longitude: req.Coordinates.Longitude,
		},
		Accuracy:   req.Accuracy,
		Heading:    req.Heading,
		Speed:      req.Speed,
		CapturedAt: req.CapturedAt,
	}

	return s.repo.Save(ctx, loc)
}

func (s *LocationService) Get(ctx context.Context, userID string) (*domain.UserLocation, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}
	return s.repo.Get(ctx, userID)
}

func (s *LocationService) RemoveUserLocation(ctx context.Context, userID string) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}
	return s.repo.RemoveUserLocation(ctx, userID)
}

func (s *LocationService) SearchNearby(ctx context.Context, req *SearchNearbyRequest) ([]*domain.UserLocation, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}
	return s.repo.SearchNearby(ctx, req.Coordinates, req.RadiusKm)
}
