package pgstore

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/config"
)

type FareConfigRepo struct {
	config *config.Config
	db     *pgstore.PostgresDB
}

func NewFareConfigRepo(config *config.Config, db *pgstore.PostgresDB) *FareConfigRepo {
	return &FareConfigRepo{
		config: config,
		db:     db,
	}
}
