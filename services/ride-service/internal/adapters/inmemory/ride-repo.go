package inmemory

import (
	"context"

	"github.com/docker/distribution/uuid"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/domain"
)

const bufferRides = 1000

type InMemoryRideRepo struct {
	data map[uuid.UUID]*domain.Ride
}

func NewInMemoryRideRepo() *InMemoryRideRepo {
	return &InMemoryRideRepo{
		data: make(map[uuid.UUID]*domain.Ride, bufferRides),
	}
}

func (repo *InMemoryRideRepo) Add(ctx context.Context, ride *domain.Ride) error {
	if _, exists := repo.data[ride.ID]; exists {
		return domain.ErrDuplicateRide
	}
	repo.data[ride.ID] = ride
	return nil
}

func (repo *InMemoryRideRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ride, error) {
	ride, exists := repo.data[id]
	if !exists {
		return nil, domain.ErrRideNotFound
	}
	return ride, nil
}
