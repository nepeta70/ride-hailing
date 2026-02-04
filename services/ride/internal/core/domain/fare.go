package domain

import (
	"time"

	"github.com/docker/distribution/uuid"
)

type FareType string

const (
	FareTypeStandard FareType = "STANDARD"
	FareTypePremium  FareType = "PREMIUM"
)

type FareRequest struct {
	RequestID       uuid.UUID
	PickupLocation  string
	DropoffLocation string
	FareType        FareType
}

type Fare struct {
	ID                  uuid.UUID
	Type                FareType
	Fare                float64
	EstimatedDistanceKm float64
	EstimatedDuration   time.Duration
	ETA                 time.Time
	Currency            string
}

type DirectionsResponse struct {

	// Distance in meters.
	Distance float64 `json:"distance"`

	// Duration indicates total time required for this leg.
	Duration time.Duration `json:"duration"`

	// DurationInTraffic indicates the total duration of this leg. This value is an
	// estimate of the time in traffic based on current and historical traffic
	// conditions.
	DurationInTraffic time.Duration `json:"duration_in_traffic"`

	// ArrivalTime contains the estimated time of arrival for this leg. This property
	// is only returned for transit directions.
	ArrivalTime time.Time `json:"arrival_time"`
}
