package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
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

type StorageBundle interface {
	FareReadRepo() FareReadRepository
	FareWriteRepo() FareWriteRepository
	FareRatesReadRepo() FareRatesReadRepository
	FareRatesWriteRepo() FareRatesWriteRepository
	CountryCache() CountryCacheInterface
	Silo() pkgPorts.Silo
	GrainStorage() GrainStorage
}

type ServiceTypeInterface interface {
	GetByCode(ctx context.Context, code string) (*domain.ServiceType, error)
	Refresh(ctx context.Context) error
}

type GrainStorage interface {
	Persist(ctx context.Context, identity *grain.GrainIdentity, data *domain.GrainData) error
	Load(ctx context.Context, identity *grain.GrainIdentity, target any) (int, error)
}

type GrainSystemInterface interface {
	Silo() pkgPorts.Silo
	GrainIdentityFactory() *service.GrainIdentityFactory
	GrainPersistence() GrainStorage
}
