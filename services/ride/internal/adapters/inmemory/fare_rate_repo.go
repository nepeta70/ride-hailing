package inmemory

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type InMemoryFareRateRepo struct {
	data map[string][]*domain.FareRate
}

func NewInMemoryFareRateRepo() *InMemoryFareRateRepo {
	return &InMemoryFareRateRepo{
		data: make(map[string][]*domain.FareRate),
	}
}

func (repo *InMemoryFareRateRepo) Save(ctx context.Context, fareRate *domain.FareRate) error {
	rates, exists := repo.data[fareRate.CountryCode]
	if !exists {
		rates = []*domain.FareRate{fareRate}
	} else {
		rates = append(rates, fareRate)
	}
	repo.data[fareRate.CountryCode] = rates

	return nil
}

func (repo *InMemoryFareRateRepo) GetRatesByCountry(ctx context.Context, countryCode string) ([]*domain.FareRate, error) {
	rates, exists := repo.data[countryCode]
	if !exists {
		return nil, nil
	}
	return rates, nil
}

var _ ports.FareRatesReadRepository = (*InMemoryFareRateRepo)(nil)
var _ ports.FareRatesWriteRepository = (*InMemoryFareRateRepo)(nil)
