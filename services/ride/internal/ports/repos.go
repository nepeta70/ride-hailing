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

type RideWriteRepository interface {
	// Define methods for writing ride data
}
type RideReadRepository interface {
	// Define methods for reading ride data
}

type StorageBundle interface {
	RideReadRepo() RideReadRepository
	RideWriteRepo() RideWriteRepository
	FareReadRepo() FareReadRepository
	FareWriteRepo() FareWriteRepository
	FareRatesReadRepo() FareRatesReadRepository
	FareRatesWriteRepo() FareRatesWriteRepository
	CountryCache() CountryCacheInterface
}

type DirectionsService interface {
	GetDirections(ctx context.Context, origin, destination string) (*domain.DirectionsResponse, error)
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

type ServiceTypeInterface interface {
	GetByCode(ctx context.Context, code string) (*domain.ServiceType, error)
	Refresh(ctx context.Context) error
}
