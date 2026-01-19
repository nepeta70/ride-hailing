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

// Update updates an existing ride in the repository.
func (repo *InMemoryRideRepo) Update(ctx context.Context, ride *domain.Ride) error {
	if _, exists := repo.data[ride.ID]; !exists {
		return domain.ErrRideNotFound
	}
	repo.data[ride.ID] = ride
	return nil
}

// Delete removes a ride from the repository by ID.
func (repo *InMemoryRideRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := repo.data[id]; !exists {
		return domain.ErrRideNotFound
	}
	delete(repo.data, id)
	return nil
}

// List returns all rides in the repository.
func (repo *InMemoryRideRepo) List(ctx context.Context) ([]*domain.Ride, error) {
	rides := make([]*domain.Ride, 0, len(repo.data))
	for _, ride := range repo.data {
		rides = append(rides, ride)
	}
	return rides, nil
}
