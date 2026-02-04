package ports

import (
	"context"

	"github.com/docker/distribution/uuid"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

type FareReadRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Fare, error)
}

type FareWriteRepository interface {
	Save(ctx context.Context, fare *domain.Fare) error
}

type FareRatesReadRepository interface {
	GetRatesByRegion(ctx context.Context, countryCode, region string) ([]*domain.FareRate, error)
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
}

type DirectionsService interface {
	// Define methods for interacting with directions services
	GetDirections(ctx context.Context, origin, destination string) (*domain.DirectionsResponse, error)
}
