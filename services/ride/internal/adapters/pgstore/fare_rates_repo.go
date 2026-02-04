package pgstore

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
)

type FareRatesRepo struct {
	config *config.Config
	db     *pgstore.PostgresDB
}

func NewFareRatesRepo(config *config.Config, db *pgstore.PostgresDB) *FareRatesRepo {
	return &FareRatesRepo{
		config: config,
		db:     db,
	}
}
