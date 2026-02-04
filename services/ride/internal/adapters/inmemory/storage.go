package inmemory

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type InMemoryStorage struct {
	rideReadRepo       ports.RideReadRepository
	rideWriteRepo      ports.RideWriteRepository
	fareReadRepo       ports.FareReadRepository
	fareWriteRepo      ports.FareWriteRepository
	fareRatesReadRepo  ports.FareRatesReadRepository
	fareRatesWriteRepo ports.FareRatesWriteRepository
}

func (sa *InMemoryStorage) RideReadRepo() ports.RideReadRepository {
	return sa.rideReadRepo
}

func (sa *InMemoryStorage) RideWriteRepo() ports.RideWriteRepository {
	return sa.rideWriteRepo
}
func (sa *InMemoryStorage) FareReadRepo() ports.FareReadRepository {
	return sa.fareReadRepo
}
func (sa *InMemoryStorage) FareWriteRepo() ports.FareWriteRepository {
	return sa.fareWriteRepo
}

func (sa *InMemoryStorage) FareRatesReadRepo() ports.FareRatesReadRepository {
	// Implement the FareRatesReadRepository retrieval logic here
	return sa.fareRatesReadRepo
}
func (sa *InMemoryStorage) FareRatesWriteRepo() ports.FareRatesWriteRepository {
	// Implement the FareRatesWriteRepository retrieval logic here
	return sa.fareRatesWriteRepo
}

func (sa *InMemoryStorage) Close() error {
	// Implement any necessary cleanup logic here
	return nil
}

func (sa *InMemoryStorage) HealthCheck(ctx context.Context) error {
	// Implement any necessary health check logic here
	return nil
}

func NewInMemoryStorageBundle() (*InMemoryStorage, error) {
	rideReadRepo := NewInMemoryRideRepo()
	rideWriteRepo := NewInMemoryRideRepo()
	fareReadRepo := NewInMemoryFareRepo()
	fareWriteRepo := NewInMemoryFareRepo()

	return &InMemoryStorage{
		rideReadRepo:  rideReadRepo,
		rideWriteRepo: rideWriteRepo,
		fareReadRepo:  fareReadRepo,
		fareWriteRepo: fareWriteRepo,
	}, nil
}

var _ ports.StorageBundle = (*InMemoryStorage)(nil)
