package config

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

type LocationServiceConfig struct {
	LocationServiceAddress string `json:"location_service_address" env:"LOCATION_SERVICE_ADDRESS"`
	APIKey                 string `json:"api_key" env:"LOCATION_SERVICE_API_KEY"`
	SenderID               string `json:"sender_id" env:"SENDER_ID"`
}

func (c *LocationServiceConfig) Validate() error {
	if c.LocationServiceAddress == "" {
		return errors.NewValidationErrorf("LocationServiceAddress is required")
	}
	if c.APIKey == "" {
		return errors.NewValidationErrorf("APIKey is required")
	}
	if c.SenderID == "" {
		return errors.NewValidationErrorf("SenderID is required")
	}
	return nil
}

func DefaultLocationServiceConfig() *LocationServiceConfig {
	return &LocationServiceConfig{
		LocationServiceAddress: "localhost:50051",
		APIKey:                 "",
		SenderID:               "",
	}
}
