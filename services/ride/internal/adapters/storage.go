package adapters

import (
	"context"

	rd "github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/inmemory"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type RedisStorage struct {
	rideReadRepo       ports.RideReadRepository
	rideWriteRepo      ports.RideWriteRepository
	fareReadRepo       ports.FareReadRepository
	fareWriteRepo      ports.FareWriteRepository
	fareRatesReadRepo  ports.FareRatesReadRepository
	fareRatesWriteRepo ports.FareRatesWriteRepository
}

func (sa *RedisStorage) RideReadRepo() ports.RideReadRepository {
	return sa.rideReadRepo
}

func (sa *RedisStorage) RideWriteRepo() ports.RideWriteRepository {
	return sa.rideWriteRepo
}
func (sa *RedisStorage) FareReadRepo() ports.FareReadRepository {
	return sa.fareReadRepo
}
func (sa *RedisStorage) FareWriteRepo() ports.FareWriteRepository {
	return sa.fareWriteRepo
}

func (sa *RedisStorage) FareRatesReadRepo() ports.FareRatesReadRepository {
	return sa.fareRatesReadRepo
}

func (sa *RedisStorage) FareRatesWriteRepo() ports.FareRatesWriteRepository {
	return sa.fareRatesWriteRepo
}

func (sa *RedisStorage) Close() error {
	// Implement any necessary cleanup logic here
	return nil
}

func (sa *RedisStorage) HealthCheck(ctx context.Context) error {
	// Implement any necessary health check logic here
	return nil
}

func NewRedisStorageBundle(cfg *config.Config, client *rd.RedisClient, logger pkgPorts.Logger) (*RedisStorage, error) {
	rideReadRepo := inmemory.NewInMemoryRideRepo()
	rideWriteRepo := inmemory.NewInMemoryRideRepo()
	fareReadRepo := rdstore.NewFareRepository(cfg, client, logger)
	fareWriteRepo := rdstore.NewFareRepository(cfg, client, logger)
	return &RedisStorage{
		rideReadRepo:  rideReadRepo,
		rideWriteRepo: rideWriteRepo,
		fareReadRepo:  fareReadRepo,
		fareWriteRepo: fareWriteRepo,
	}, nil
}

var _ ports.StorageBundle = (*RedisStorage)(nil)
