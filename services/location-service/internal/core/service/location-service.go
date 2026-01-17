// internal/location/service/service.go
package service

import (
	"context"

	"github.com/mmcloughlin/geohash"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/ports"
)

type LocationService struct {
	repo ports.LocationRepository
}

func NewLocationService(r ports.LocationRepository) *LocationService {
	return &LocationService{repo: r}
}

func (s *LocationService) Update(ctx context.Context, req *UpdateRequest) error {
	if err := req.Validate(); err != nil {
		return err // Returns the Business Error
	}
	// 1. Calculate Geohash (Logic stays in the Service)
	hash := geohash.Encode(req.Latitude, req.Longitude)

	// 2. Map to Domain (The Geohash-only object you wanted)
	loc := &domain.Location{
		EntityID:   req.EntityID,
		Geohash:    hash,
		Accuracy:   req.Accuracy,
		Heading:    req.Heading,
		Speed:      req.Speed,
		CapturedAt: req.CapturedAt,
	}

	return s.repo.Save(ctx, loc)
}

func (s *LocationService) Get(ctx context.Context, entityID string) (*domain.Location, error) {
	return s.repo.Get(ctx, entityID)
}
