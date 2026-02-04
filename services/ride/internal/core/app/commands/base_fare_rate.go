package commands

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

type FareRateData struct {
	CountryCode   string
	RegionCode    string
	ServiceType   string
	BaseFare      float64
	FarePerKm     float64
	FarePerMinute float64
	MinimumFare   float64
	Currency      string
}

func (c *FareRateData) Validate() error {

	if c.CountryCode == "" || c.RegionCode == "" || c.ServiceType == "" {
		return errors.NewValidationErrorf("country code, region code, and service type are required")
	}
	if len(c.CountryCode) != 2 {
		return errors.NewValidationErrorf("country code must be 2 characters")
	}
	if len(c.RegionCode) != 2 {
		return errors.NewValidationErrorf("region code must be 2 characters")
	}
	if len(c.ServiceType) == 0 {
		return errors.NewValidationErrorf("service type must be provided")
	}
	if c.BaseFare < 0 || c.FarePerKm < 0 || c.FarePerMinute < 0 || c.MinimumFare < 0 {
		return errors.NewValidationErrorf("fare components must be non-negative")
	}
	return nil
}
