package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

type FareReadRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Fares, error)
}

type FareWriteRepository interface {
	Save(ctx context.Context, fare *domain.Fares) error
}

type FareRatesReadRepository interface {
	GetRatesByCountry(ctx context.Context, countryCode string) ([]*domain.FareRate, error)
}

type FareRatesWriteRepository interface {
	Save(ctx context.Context, fareRate *domain.FareRate) error
}

type CountryReadRepoInterface interface {
	GetAllEnabled(ctx context.Context) (map[string]*domain.Country, error)
}

type CountryCacheInterface interface {
	GetCountryByCode(ctx context.Context, code string) (*domain.Country, bool)
	Refresh(ctx context.Context) error
}

type ServiceTypeReadRepository interface {
	GetAllEnabled(ctx context.Context) (map[string]*domain.ServiceType, error)
}

type ServiceTypeCacheInterface interface {
	GetServiceTypeByCode(ctx context.Context, code string) (*domain.ServiceType, bool)
	Refresh(ctx context.Context) error
}

type StorageBundle interface {
	FareReadRepo() FareReadRepository
	FareWriteRepo() FareWriteRepository
	FareRatesReadRepo() FareRatesReadRepository
	FareRatesWriteRepo() FareRatesWriteRepository
	CountryCache() CountryCacheInterface
	ServiceTypeCache() ServiceTypeCacheInterface
	GrainStorage() GrainStorage
}
