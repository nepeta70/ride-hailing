package config

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type FareConfig struct {
	DefaultCurrency string  `json:"default_currency" env:"RIDE_DEFAULT_CURRENCY"`
	BaseFare        float64 `json:"base_fare" env:"RIDE_BASE_FARE"`
	FarePerKm       float64 `json:"fare_per_km" env:"RIDE_FARE_PER_KM"`
	FarePerMinute   float64 `json:"fare_per_minute" env:"RIDE_FARE_PER_MINUTE"`
}

func (rc *FareConfig) Validate() error {
	// Add validation logic if needed
	if rc.BaseFare <= 0.0 {
		return errors.NewValidationErrorf("BaseFare must be greater than 0")
	}
	if rc.FarePerKm <= 0.0 {
		return errors.NewValidationErrorf("FarePerKm must be greater than 0")
	}
	if rc.FarePerMinute <= 0.0 {
		return errors.NewValidationErrorf("FarePerMinute must be greater than 0")
	}

	return nil
}

func DefaultFareConfig() FareConfig {
	return FareConfig{
		DefaultCurrency: "EUR",
		BaseFare:        3.0,
		FarePerKm:       1.0,
		FarePerMinute:   0.5,
	}
}
