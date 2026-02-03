package pgstore

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/postgres"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/config"
)

type FareConfigRepo struct {
	config *config.Config
	db     *postgres.PostgresDB
}

func NewFareConfigRepo(config *config.Config, db *postgres.PostgresDB) *FareConfigRepo {
	return &FareConfigRepo{
		config: config,
		db:     db,
	}
}
