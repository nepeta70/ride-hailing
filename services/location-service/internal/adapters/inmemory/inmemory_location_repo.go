package inmemory

import (
	"context"

	"github.com/mmcloughlin/geohash"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
)

// InMemoryLocationRepo is an in-memory implementation of the LocationRepository interface.
type InMemoryLocationRepo struct {
	store map[string]*domain.UserLocation
	index map[string][]string
}

func NewInMemoryLocationRepo() *InMemoryLocationRepo {
	return &InMemoryLocationRepo{
		store: make(map[string]*domain.UserLocation),
		index: make(map[string][]string),
	}
}

func (r *InMemoryLocationRepo) Save(ctx context.Context, loc *domain.UserLocation) error {
	r.store[loc.UserID] = loc
	if loc.UserType == domain.UserTypeDriver {
		hash := geohash.EncodeWithPrecision(loc.Coordinates.Latitude, loc.Coordinates.Longitude, 8)
		r.index[hash] = append(r.index[hash], loc.UserID)
	}
	return nil
}

func (r *InMemoryLocationRepo) Get(ctx context.Context, userID string) (*domain.UserLocation, error) {
	loc, ok := r.store[userID]
	if !ok {
		return nil, domain.ErrLocationNotFound
	}
	return loc, nil
}

func (r *InMemoryLocationRepo) RemoveUserLocation(ctx context.Context, userID string) error {
	delete(r.store, userID)
	return nil
}

func (r *InMemoryLocationRepo) SearchNearby(ctx context.Context, coordinates domain.Coordinates, radiusKm float32) ([]*domain.UserLocation, error) {
	var results []*domain.UserLocation
	hash := geohash.EncodeWithPrecision(coordinates.Latitude, coordinates.Longitude, 8)
	userIDs, ok := r.index[hash]
	for _, userID := range userIDs {
		if loc, exists := r.store[userID]; exists {
			results = append(results, loc)
		}
	}
	if !ok {
		return results, nil
	}
	return results, nil
}
