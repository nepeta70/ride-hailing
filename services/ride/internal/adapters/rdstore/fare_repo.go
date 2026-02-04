package rdstore

import (
	"context"
	"encoding/json"

	"github.com/docker/distribution/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/rdstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
	"github.com/redis/go-redis/v9"
)

const fareKeyPrefix = "fare:"

type FareRepository struct {
	client *redis.Client
	cfg    *config.Config
	logger pkgPorts.Logger
}

func NewFareRepository(cfg *config.Config, client *rdstore.RedisClient, logger pkgPorts.Logger) *FareRepository {
	return &FareRepository{
		cfg:    cfg,
		client: client.Rdb,
		logger: logger,
	}
}

func (r *FareRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Fare, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()

	fareKey := fareKey(id)
	data, err := r.client.HGet(ctx, fareKey, "data").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.NewErrNotFound("Fare not found")
		}
		return nil, err
	}

	var fare domain.Fare
	err = json.Unmarshal([]byte(data), &fare)
	if err != nil {
		return nil, errors.NewErrJSONUnmarshal(err)
	}

	return &fare, nil
}

func (r *FareRepository) Save(ctx context.Context, fare *domain.Fare) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeouts.RequestTimeout)
	defer cancel()

	fareKey := fareKey(fare.ID)

	data, err := json.Marshal(fare)
	if err != nil {
		return errors.NewErrJSONMarshal(err)
	}

	err = r.client.HSet(ctx, fareKey, "data", data).Err()
	if err != nil {
		return err
	}
	return nil
}

var _ ports.FareReadRepository = (*FareRepository)(nil)
var _ ports.FareWriteRepository = (*FareRepository)(nil)

func fareKey(fareID uuid.UUID) string {
	return fareKeyPrefix + fareID.String()
}
