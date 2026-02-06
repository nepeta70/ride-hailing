package pgstore

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/shopspring/decimal"
)

type FareRateReadEntity struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Country     string    `db:"country_code" json:"country_code"`
	ServiceType string    `db:"service_type" json:"service_type"`

	// Use decimal.Decimal to match Postgres NUMERIC(12,2)
	BaseFare      decimal.Decimal `db:"base_fare" json:"base_fare"`
	CostPerKm     decimal.Decimal `db:"cost_per_km" json:"cost_per_km"`
	CostPerMinute decimal.Decimal `db:"cost_per_minute" json:"cost_per_minute"`
	MinimumFare   decimal.Decimal `db:"minimum_fare" json:"minimum_fare"`
}

func (e FareRateReadEntity) ToDomain(currency string) domain.FareRate {
	return domain.FareRate{
		ID:            e.ID,
		BaseFare:      e.BaseFare.InexactFloat64(),
		FarePerKm:     e.CostPerKm.InexactFloat64(),
		FarePerMinute: e.CostPerMinute.InexactFloat64(),
		MinimumFare:   e.MinimumFare.InexactFloat64(),
		Currency:      currency, // Passed from the countries table/service
		CountryCode:   e.Country,
		ServiceType:   e.ServiceType,
	}
}

// FareRateEntity represents the database row for fare configurations.
// This is private to the adapter and only used for SQL mapping.
type FareRateWriteEntity struct {
	FareRateReadEntity
	IsActive  bool   `db:"is_active" json:"is_active"`
	Version   int    `db:"version" json:"version"`
	VersionBy string `db:"version_by" json:"version_by"`
}
