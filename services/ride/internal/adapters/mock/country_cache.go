package mock

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type MockCountryCache struct {
	countries map[string]*domain.Country
}

func (m *MockCountryCache) GetCountryByCode(ctx context.Context, code string) (*domain.Country, bool) {
	country, exists := m.countries[code]
	return country, exists
}

func (m *MockCountryCache) Refresh(ctx context.Context) error {
	// No-op for mock
	return nil
}

func NewMockCountryCache() *MockCountryCache {
	return &MockCountryCache{
		countries: map[string]*domain.Country{
			"ES": {Code: "ES", Currency: "EUR"}, // Spain
			"FR": {Code: "FR", Currency: "EUR"}, // France
			"DE": {Code: "DE", Currency: "EUR"}, // Germany
			"GB": {Code: "GB", Currency: "GBP"}, // United Kingdom
			"PL": {Code: "PL", Currency: "PLN"}, // Poland
		},
	}
}

var _ ports.CountryCacheInterface = (*MockCountryCache)(nil)
