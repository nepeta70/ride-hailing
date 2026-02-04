package domain

import "github.com/docker/distribution/uuid"

type FareRate struct {
	ID            uuid.UUID
	BaseFare      float64
	FarePerKm     float64
	FarePerMinute float64
	MinimumFare   float64
	Currency      string
	CountryCode   string
	RegionCode    string
	ServiceType   string
}


