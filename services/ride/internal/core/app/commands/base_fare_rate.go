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
	if len(c.CountryCode) != 2 {
		return errors.NewValidationErrorf("country code must have 2 characters")
	}

	if len(c.RegionCode) != 2 && len(c.RegionCode) != 0 {
		return errors.NewValidationErrorf("region code must have 2 characters or be empty")
	}

	if len(c.ServiceType) == 0 {
		return errors.NewValidationErrorf("service type must be provided")
	}

	if c.BaseFare < 0 || c.FarePerKm < 0 || c.FarePerMinute < 0 || c.MinimumFare < 0 {
		return errors.NewValidationErrorf("fare components must be non-negative")
	}
	return nil
}
