package cache

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type CountryCache struct {
	countryReadRepo ports.CountryReadRepoInterface
	store           atomic.Pointer[map[string]*domain.Country]
	once            sync.Once
}

func NewCountryCache(countryReadRepo ports.CountryReadRepoInterface) *CountryCache {
	return &CountryCache{
		countryReadRepo: countryReadRepo,
	}
}

func (c *CountryCache) GetCountryByCode(ctx context.Context, code string) (*domain.Country, bool) {
	c.once.Do(func() {
		_ = c.Refresh(ctx)
	})

	data := c.store.Load()
	if data == nil {
		return nil, false
	}

	country, ok := (*data)[code]
	return country, ok
}

func (c *CountryCache) Refresh(ctx context.Context) error {
	store, err := c.countryReadRepo.GetAllEnabled(ctx)
	if err != nil {
		return err
	}

	c.store.Store(&store)
	return nil
}

var _ ports.CountryCacheInterface = (*CountryCache)(nil)
