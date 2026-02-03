package pgstore

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FareConfigEntity represents the database row for fare configurations.
// This is private to the adapter and only used for SQL mapping.
type FareConfigEntity struct {
	ID           uuid.UUID `db:"id" json:"id"`
	RegionCode   string    `db:"region_code" json:"region_code"`
	ServiceType  string    `db:"service_type" json:"service_type"`
	CurrencyCode string    `db:"currency_code" json:"currency_code"`

	// Use decimal.Decimal to match Postgres NUMERIC(12,2)
	BaseFare      decimal.Decimal `db:"base_fare" json:"base_fare"`
	CostPerKm     decimal.Decimal `db:"cost_per_km" json:"cost_per_km"`
	CostPerMinute decimal.Decimal `db:"cost_per_minute" json:"cost_per_minute"`
	MinimumFare   decimal.Decimal `db:"minimum_fare" json:"minimum_fare"`

	IsActive bool `db:"is_active" json:"is_active"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
