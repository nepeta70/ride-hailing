package cache

import (
	"context"
	"sync"
	"sync/atomic"

	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type CountryCache struct {
	countryReadRepo ports.CountryReadRepoInterface
	store           atomic.Pointer[map[string]*domain.Country]
	once            sync.Once
	logger          pkgPorts.Logger
}

func NewCountryCache(countryReadRepo ports.CountryReadRepoInterface, logger pkgPorts.Logger) *CountryCache {
	return &CountryCache{
		countryReadRepo: countryReadRepo,
		logger:          logger,
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

	c.logger.Info("Refreshing country cache", "entries", len(store))
	c.store.Store(&store)
	return nil
}

var _ ports.CountryCacheInterface = (*CountryCache)(nil)
