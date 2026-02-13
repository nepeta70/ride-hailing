package adapters

import (
	"context"

	pg "github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	rd "github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/cache"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/inmemory"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/mock"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type StorageBundleOptions struct {
	Config   *config.Config
	RdClient *rd.RedisClient
	PgDB     *pg.PostgresDB
	Logger   pkgPorts.Logger
}

func (opts *StorageBundleOptions) Validate() error {
	if opts.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if opts.RdClient == nil {
		return errors.NewValidationErrorf("redis client is required")
	}
	if opts.PgDB == nil {
		return errors.NewValidationErrorf("postgres db is required")
	}
	if opts.Logger == nil {
		return errors.NewValidationErrorf("logger is required")
	}
	return nil
}

type StorageBundle struct {
	fareReadRepo       ports.FareReadRepository
	fareWriteRepo      ports.FareWriteRepository
	fareRatesReadRepo  ports.FareRatesReadRepository
	fareRatesWriteRepo ports.FareRatesWriteRepository
	countryCache       ports.CountryCacheInterface
	grainStorage       ports.GrainStorage
	serviceTypeCache   ports.ServiceTypeCacheInterface
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

func (sa *StorageBundle) GrainStorage() ports.GrainStorage {
	return sa.grainStorage
}

func (sa *StorageBundle) FareRatesWriteRepo() ports.FareRatesWriteRepository {
	return sa.fareRatesWriteRepo
}

func (sa *StorageBundle) ServiceTypeCache() ports.ServiceTypeCacheInterface {
	return sa.serviceTypeCache
}

func (sa *StorageBundle) Close() error {
	// Implement any necessary cleanup logic here
	return nil
}

func (sa *StorageBundle) HealthCheck(ctx context.Context) error {
	// Implement any necessary health check logic here
	return nil
}

func NewRedisStorageBundle(opts *StorageBundleOptions) (*StorageBundle, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	countryRepo := pgstore.NewCountryReadRepo(opts.Config, opts.PgDB)
	countryCache := cache.NewCountryCache(countryRepo, opts.Logger)
	fareRepo := rdstore.NewFareRepository(opts.Config, opts.RdClient, opts.Logger)
	serviceTypeRepo := pgstore.NewServiceTypeReadRepo(opts.Config, opts.PgDB)
	serviceTypeCache := cache.NewServiceTypeCache(serviceTypeRepo)

	return &StorageBundle{
		fareReadRepo:       fareRepo,
		fareWriteRepo:      fareRepo,
		fareRatesReadRepo:  mock.NewMockFareRatesRepo(),
		fareRatesWriteRepo: inmemory.NewInMemoryFareRateRepo(),
		countryCache:       countryCache,
		grainStorage:       pgstore.NewGrainStorage(opts.PgDB),
		serviceTypeCache:   serviceTypeCache,
	}, nil
}

var _ ports.StorageBundle = (*StorageBundle)(nil)
