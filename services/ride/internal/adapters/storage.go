package adapters

import (
	"context"

	pg "github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	rd "github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/cache"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/inmemory"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/mock"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type StorageBundle struct {
	rideReadRepo       ports.RideReadRepository
	rideWriteRepo      ports.RideWriteRepository
	fareReadRepo       ports.FareReadRepository
	fareWriteRepo      ports.FareWriteRepository
	fareRatesReadRepo  ports.FareRatesReadRepository
	fareRatesWriteRepo ports.FareRatesWriteRepository
	countryCache       ports.CountryCacheInterface
}

func (sa *StorageBundle) RideReadRepo() ports.RideReadRepository {
	return sa.rideReadRepo
}

func (sa *StorageBundle) RideWriteRepo() ports.RideWriteRepository {
	return sa.rideWriteRepo
}
func (sa *StorageBundle) FareReadRepo() ports.FareReadRepository {
	return sa.fareReadRepo
}
func (sa *StorageBundle) FareWriteRepo() ports.FareWriteRepository {
	return sa.fareWriteRepo
}

func (sa *StorageBundle) FareRatesReadRepo() ports.FareRatesReadRepository {
	return sa.fareRatesReadRepo
}

func (sa *StorageBundle) CountryCache() ports.CountryCacheInterface {
	return sa.countryCache
}

func (sa *StorageBundle) FareRatesWriteRepo() ports.FareRatesWriteRepository {
	return sa.fareRatesWriteRepo
}

func (sa *StorageBundle) Close() error {
	// Implement any necessary cleanup logic here
	return nil
}

func (sa *StorageBundle) HealthCheck(ctx context.Context) error {
	// Implement any necessary health check logic here
	return nil
}

func NewRedisStorageBundle(cfg *config.Config, rdclient *rd.RedisClient, pgdb *pg.PostgresDB, logger pkgPorts.Logger) (*StorageBundle, error) {
	countryRepo := pgstore.NewCountryReadRepo(cfg, pgdb)
	countryCache := cache.NewCountryCache(countryRepo)
	rideReadRepo := inmemory.NewInMemoryRideRepo()
	rideWriteRepo := inmemory.NewInMemoryRideRepo()
	fareReadRepo := rdstore.NewFareRepository(cfg, rdclient, logger)
	fareWriteRepo := rdstore.NewFareRepository(cfg, rdclient, logger)
	return &StorageBundle{
		rideReadRepo:       rideReadRepo,
		rideWriteRepo:      rideWriteRepo,
		fareReadRepo:       fareReadRepo,
		fareWriteRepo:      fareWriteRepo,
		fareRatesReadRepo:  mock.NewMockFareRatesRepo(),
		fareRatesWriteRepo: inmemory.NewInMemoryFareRateRepo(),
		countryCache:       countryCache,
	}, nil
}

var _ ports.StorageBundle = (*StorageBundle)(nil)
