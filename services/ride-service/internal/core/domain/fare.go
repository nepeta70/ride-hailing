package domain

import "github.com/docker/distribution/uuid"

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
	ID                   uuid.UUID
	Type                 FareType
	Fare                 float64
	EstimatedDistanceKm  float64
	EstimatedDurationMin float64
}
