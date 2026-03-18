package config

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

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

type KeysConfig struct {
	GoogleMapsAPIKey string `json:"google_maps_api_key" env:"GOOGLE_MAPS_API_KEY"`
}

func (rc *KeysConfig) Validate() error {

	return nil
}

func DefaultKeysConfig() KeysConfig {
	return KeysConfig{
		GoogleMapsAPIKey: "",
	}
}

type RideConfig struct {
	RideRequestTimeoutMinutes int           `json:"ride_request_timeout" env:"RIDE_REQUEST_TIMEOUT"`
	RideRequestTimeout        time.Duration `json:"-" env:"-"`
}

func (rc *RideConfig) Validate() error {
	return nil
}

func (rc *RideConfig) Init() error {
	rc.RideRequestTimeout = time.Duration(rc.RideRequestTimeoutMinutes) * time.Minute
	return nil
}

func DefaultRideConfig() RideConfig {
	cfg := RideConfig{
		RideRequestTimeoutMinutes: 5,
	}
	cfg.Init()
	return cfg
}
