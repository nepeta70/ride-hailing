package domain

import (
	"time"

	"github.com/google/uuid"
)

type Fares struct {
	RequestID                uuid.UUID
	ID                       uuid.UUID
	PickupLocation           string
	DropoffLocation          string
	EstimatedDistanceKm      float64
	EstimatedDurationMinutes time.Duration
	ETA                      time.Time
	Currency                 string
	Timestamp                time.Time
	Fares                    map[string]float64 // service type -> fare
}

type DirectionsResponse struct {

	// DistanceMeters in meters.
	DistanceMeters float64 `json:"distance"`

	// DurationMinutes indicates total time required for this leg.
	DurationMinutes time.Duration `json:"duration_in_minutes"`

	// DurationInTraffic indicates the total duration of this leg. This value is an
	// estimate of the time in traffic based on current and historical traffic
	// conditions.
	DurationInTraffic time.Duration `json:"duration_in_traffic"`

	// ArrivalTime contains the estimated time of arrival for this leg. This property
	// is only returned for transit directions.
	ArrivalTime time.Time `json:"arrival_time"`

	// Region is the region code for this leg, specified as a ccTLD two-character value.
	Region string `json:"region"`
}
