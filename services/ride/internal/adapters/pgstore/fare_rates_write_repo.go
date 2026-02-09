package pgstore

// import (
// 	"context"
// 	"time"

// 	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
// 	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
// 	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
// 	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
// )

// const (
// 	fareRatesCacheKeyPrefix = "fare_rates:"
// )

// var (
// 	fareRatesCacheExpiry = time.Hour * 24
// )

// type FareRatesRepo struct {
// 	config *config.Config
// 	db     *pgstore.PostgresDB
// 	cache  ports.CacheService
// }

// func NewFareRatesRepo(config *config.Config, db *pgstore.PostgresDB, cache ports.CacheService) *FareRatesRepo {
// 	return &FareRatesRepo{
// 		config: config,
// 		db:     db,
// 		cache:  cache,
// 	}
// }

// func (r *FareRatesRepo) GetRatesByCountry(ctx context.Context, country string) ([]*domain.FareRate, error) {
// 	var fareRates []FareRateReadEntity
// 	err := r.cache.GetOrSet(ctx, fareRatesCacheKeyPrefix+country, fareRatesCacheExpiry, fareRates, func() (any, error) {
// 		query := `
//             SELECT 
//                 id, country_code, service_type_id, 
//                 base_fare, cost_per_km, cost_per_minute, minimum_fare
//             FROM fare_rates
//             WHERE country_code = $1 
//               AND is_active = true;
//         `
// 		rows, err := r.db.QueryContext(ctx, query, country)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer rows.Close()

// 		for rows.Next() {
// 			var f FareRateReadEntity
// 			err := rows.Scan(
// 				&f.ID, &f.Country, &f.ServiceType,
// 				&f.BaseFare, &f.CostPerKm, &f.CostPerMinute, &f.MinimumFare,
// 			)
// 			if err != nil {
// 				return nil, err
// 			}
// 			fareRates = append(fareRates, f)
// 		}

// 		return fareRates, nil
// 	})

// 	resp := make([]*domain.FareRate, 0, len(fareRates))
// 	for _, rate := range fareRates {
// 		resp = append(resp, &domain.FareRate{
// 			ID:            rate.ID,
// 			CountryCode:   rate.Country,
// 			ServiceType:   rate.ServiceType,
// 			BaseFare:      rate.BaseFare.InexactFloat64(),
// 			FarePerKm:     rate.CostPerKm.InexactFloat64(),
// 			FarePerMinute: rate.CostPerMinute.InexactFloat64(),
// 			MinimumFare:   rate.MinimumFare.InexactFloat64(),
// 		})
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return resp, nil
// }
