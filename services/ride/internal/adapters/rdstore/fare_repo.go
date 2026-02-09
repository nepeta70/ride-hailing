package rdstore

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

const fareKeyPrefix = "fare:"

var ttl = 5 * time.Minute

type FareRepository struct {
	redisDb *rdstore.RedisDb[domain.Fares]
	cfg     *config.Config
	logger  pkgPorts.Logger
}

func NewFareRepository(cfg *config.Config, client *rdstore.RedisClient, logger pkgPorts.Logger) *FareRepository {
	redisDb := rdstore.NewRedisStorage[domain.Fares](&cfg.BaseConfig, client, logger)
	return &FareRepository{
		cfg:     cfg,
		redisDb: redisDb,
		logger:  logger,
	}
}

func (r *FareRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Fares, error) {

	fareKey := fareKey(id)
	return r.redisDb.Get(ctx, fareKey)
}

func (r *FareRepository) Save(ctx context.Context, fare *domain.Fares) error {
	return r.redisDb.Save(ctx, fareKey(fare.ID), fare, ttl)
}

var _ ports.FareReadRepository = (*FareRepository)(nil)
var _ ports.FareWriteRepository = (*FareRepository)(nil)

func fareKey(fareID uuid.UUID) string {
	return fareKeyPrefix + fareID.String()
}
