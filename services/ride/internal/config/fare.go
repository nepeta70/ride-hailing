package config

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

type FareConfig struct {
	DefaultCurrency string `json:"default_currency" env:"RIDE_DEFAULT_CURRENCY"`
}

func (rc *FareConfig) Validate() error {
	// Add validation logic if needed
	if len(rc.DefaultCurrency) != 3 {
		return errors.NewValidationErrorf("default currency must be a 3-letter code")
	}
	return nil
}

func DefaultFareConfig() FareConfig {
	return FareConfig{
		DefaultCurrency: "EUR",
	}
}
