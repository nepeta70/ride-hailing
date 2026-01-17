package inmemory

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
)

// InMemoryLocationRepo is an in-memory implementation of the LocationRepository interface.
type InMemoryLocationRepo struct {
	store map[string]*domain.Location
}

func NewInMemoryLocationRepo() *InMemoryLocationRepo {
	return &InMemoryLocationRepo{
		store: make(map[string]*domain.Location),
	}
}

func (r *InMemoryLocationRepo) Save(ctx context.Context, loc *domain.Location) error {
	r.store[loc.EntityID] = loc
	return nil
}

func (r *InMemoryLocationRepo) Get(ctx context.Context, entityID string) (*domain.Location, error) {
	loc, ok := r.store[entityID]
	if !ok {
		return nil, domain.ErrLocationNotFound
	}
	return loc, nil
}
