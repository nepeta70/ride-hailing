package pgstore

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

type FareRatesReadRepo struct {
	config *config.Config
	db     *pgstore.PostgresDB
}

func NewFareRatesRepo(config *config.Config, db *pgstore.PostgresDB) *FareRatesReadRepo {
	return &FareRatesReadRepo{
		config: config,
		db:     db,
	}
}

func (r *FareRatesReadRepo) GetRatesByCountry(ctx context.Context, country string) ([]*domain.FareRate, error) {
	query := `
            SELECT 
                id, country_code, service_type, 
                base_fare, cost_per_km, cost_per_minute, minimum_fare
            FROM fare_rates
            WHERE country_code = $1 
              AND is_active = true;
        `
	rows, err := r.db.QueryContext(ctx, query, country)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fareRates := make([]*domain.FareRate, 0, 10)
	for rows.Next() {
		var f FareRateReadEntity
		err := rows.Scan(
			&f.ID, &f.Country, &f.ServiceType,
			&f.BaseFare, &f.CostPerKm, &f.CostPerMinute, &f.MinimumFare,
		)
		if err != nil {
			return nil, err
		}
		rate := &domain.FareRate{
			ID:            f.ID,
			CountryCode:   f.Country,
			ServiceType:   f.ServiceType,
			BaseFare:      f.BaseFare.InexactFloat64(),
			FarePerKm:     f.CostPerKm.InexactFloat64(),
			FarePerMinute: f.CostPerMinute.InexactFloat64(),
			MinimumFare:   f.MinimumFare.InexactFloat64(),
		}
		fareRates = append(fareRates, rate)
	}
	return fareRates, nil
}
