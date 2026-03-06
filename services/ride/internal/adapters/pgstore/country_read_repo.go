package pgstore

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type CountryReadEntity struct {
	Code         string `db:"code" json:"code"`
	CurrencyCode string `db:"currency_code" json:"currency_code"`
}

type CountryReadRepo struct {
	config *config.Config
	db     *pgstore.PostgresDB
}

func NewCountryReadRepo(config *config.Config, db *pgstore.PostgresDB) *CountryReadRepo {
	return &CountryReadRepo{
		config: config,
		db:     db,
	}
}

func (r *CountryReadRepo) GetAllEnabled(ctx context.Context) (map[string]*domain.Country, error) {
	query := `SELECT code, currency_code FROM countries WHERE is_enabled = true`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*domain.Country, 50)
	for rows.Next() {
		var c CountryReadEntity
		err := rows.Scan(&c.Code, &c.CurrencyCode)
		if err != nil {
			return nil, err
		}
		results[c.Code] = &domain.Country{
			Code:     c.Code,
			Currency: c.CurrencyCode,
		}
	}

	return results, nil
}

var _ ports.CountryReadRepoInterface = (*CountryReadRepo)(nil)
