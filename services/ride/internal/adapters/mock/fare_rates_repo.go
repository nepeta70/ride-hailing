package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type MockFareRatesRepo struct {
	rates map[string][]*domain.FareRate
}

func NewMockFareRatesRepo() *MockFareRatesRepo {
	return &MockFareRatesRepo{
		rates: map[string][]*domain.FareRate{
			"ES": {
				{
					ID:            uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					BaseFare:      2.50,
					FarePerKm:     1.10,
					FarePerMinute: 0.20,
					MinimumFare:   5.00,
					Currency:      "EUR",
					CountryCode:   "ES",
					ServiceType:   "STANDARD",
				},
			},
			"FR": {
				{
					ID:            uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
					BaseFare:      3.00,
					FarePerKm:     1.20,
					FarePerMinute: 0.25,
					MinimumFare:   6.00,
					Currency:      "EUR",
					CountryCode:   "FR",
					ServiceType:   "STANDARD",
				},
			},
			"DE": {
				{
					ID:            uuid.MustParse("8dbad308-0118-409c-8592-36c1d76378e6"),
					BaseFare:      3.20,
					FarePerKm:     1.50,
					FarePerMinute: 0.30,
					MinimumFare:   7.00,
					Currency:      "EUR",
					CountryCode:   "DE",
					ServiceType:   "STANDARD",
				},
			},
		},
	}
}

func (m *MockFareRatesRepo) GetRatesByCountry(ctx context.Context, countryCode string) ([]*domain.FareRate, error) {
	rates, exists := m.rates[countryCode]
	if !exists {
		return []*domain.FareRate{}, nil
	}
	return rates, nil
}

var _ ports.FareRatesReadRepository = (*MockFareRatesRepo)(nil)
