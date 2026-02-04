package inmemory

import (
	"context"

	"github.com/docker/distribution/uuid"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type InMemoryFareConfigRepo struct {
	data map[uuid.UUID]*domain.FareRate
}

func NewInMemoryFareConfigRepo() *InMemoryFareConfigRepo {
	return &InMemoryFareConfigRepo{
		data: make(map[uuid.UUID]*domain.FareRate),
	}
}

func (repo *InMemoryFareConfigRepo) Save(ctx context.Context, fareRate *domain.FareRate) error {
	repo.data[fareRate.ID] = fareRate
	return nil
}

func (repo *InMemoryFareConfigRepo) GetRatesByRegion(ctx context.Context, countryCode, region string) ([]*domain.FareRate, error) {
	var rates []*domain.FareRate
	for _, rate := range repo.data {
		if rate.CountryCode == countryCode && rate.RegionCode == region {
			rates = append(rates, rate)
		}
	}
	return rates, nil
}

var _ ports.FareRatesReadRepository = (*InMemoryFareConfigRepo)(nil)
var _ ports.FareRatesWriteRepository = (*InMemoryFareConfigRepo)(nil)
