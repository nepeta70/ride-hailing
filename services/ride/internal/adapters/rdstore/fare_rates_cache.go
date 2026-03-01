package rdstore

import (
	"context"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	ports "github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

const (
	fareRatesCacheKeyPrefix = "fare_rates:"
)

var (
	fareRatesCacheExpiry = time.Hour * 24
)

type FareRatesCache struct {
	config *config.Config
	repo   ports.FareRatesReadRepository
	cache  pkgPorts.GenericCache[[]*domain.FareRate]
	redis  *rdstore.RedisClient
}

func NewFareRatesCache(config *config.Config, repo ports.FareRatesReadRepository, redis *rdstore.RedisClient) (*FareRatesCache, error) {
	fareRatesCache, err := rdstore.NewGenericCache[[]*domain.FareRate](redis)
	if err != nil {
		return nil, err
	}
	return &FareRatesCache{
		config: config,
		repo:   repo,
		cache:  fareRatesCache,
		redis:  redis,
	}, nil
}

func (r *FareRatesCache) GetRatesByCountry(ctx context.Context, country string) ([]*domain.FareRate, error) {
	val, err := r.cache.GetOrSet(ctx, fareRatesCacheKeyPrefix+country, fareRatesCacheExpiry, func(ctx context.Context) ([]*domain.FareRate, error) {
		return r.repo.GetRatesByCountry(ctx, country)
	})

	if err != nil {
		return nil, err
	}

	return val, nil
}
