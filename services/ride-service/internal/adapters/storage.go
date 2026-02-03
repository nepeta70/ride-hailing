package adapters

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/redis"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/adapters/inmemory"
	redisAdapters "github.com/nepeta70/ride-hailing/services/ride-service/internal/adapters/redis"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/ports"
)

type RedisStorage struct {
	rideReadRepo  ports.RideReadRepository
	rideWriteRepo ports.RideWriteRepository
	fareReadRepo  ports.FareReadRepository
	fareWriteRepo ports.FareWriteRepository
}

func newRedisStorage(
	rideReadRepo ports.RideReadRepository,
	rideWriteRepo ports.RideWriteRepository,
	fareReadRepo ports.FareReadRepository,
	fareWriteRepo ports.FareWriteRepository,
) *RedisStorage {
	return &RedisStorage{
		rideReadRepo:  rideReadRepo,
		rideWriteRepo: rideWriteRepo,
		fareReadRepo:  fareReadRepo,
		fareWriteRepo: fareWriteRepo,
	}
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

func (sa *RedisStorage) Close() error {
	// Implement any necessary cleanup logic here
	return nil
}

func (sa *RedisStorage) HealthCheck(ctx context.Context) error {
	// Implement any necessary health check logic here
	return nil
}

func NewRedisStorageBundle(cfg *config.Config, client *redis.RedisClient, logger pkgPorts.Logger) (*RedisStorage, error) {
	rideReadRepo := inmemory.NewInMemoryRideRepo()
	rideWriteRepo := inmemory.NewInMemoryRideRepo()
	fareReadRepo := redisAdapters.NewFareRepository(cfg, client, logger)
	fareWriteRepo := redisAdapters.NewFareRepository(cfg, client, logger)
	return newRedisStorage(rideReadRepo, rideWriteRepo, fareReadRepo, fareWriteRepo), nil
}

var _ ports.StorageBundle = (*RedisStorage)(nil)
