package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	common "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
)

type DriverLocation struct {
	UserID          uuid.UUID              `json:"user_id"`
	SenderType      enums.SenderType       `json:"sender_type"`
	Coordinates     common.Coordinates     `json:"coordinates"`
	Accuracy        float32                `json:"accuracy"`
	Heading         float32                `json:"heading"`
	Speed           float32                `json:"speed"`
	CapturedAt      time.Time              `json:"captured_at"`
	Status          contracts.DriverStatus `json:"status"`
	StatusUpdatedAt time.Time              `json:"status_updated_at"`
	DistanceKm      float32                `json:"distance_km"`
}

func (l *DriverLocation) NewLocation() *DriverLocation {
	return &DriverLocation{}
}

type DirectionsResponse struct {
	// Distance in meters.
	Distance int `json:"distance"`

	// Duration indicates total time required for this leg.
	Duration time.Duration `json:"duration"`
}

type DriverStatusUpdate struct {
	DriverID        uuid.UUID              `json:"driver_id"`
	Status          contracts.DriverStatus `json:"status"`
	StatusUpdatedAt time.Time              `json:"status_updated_at"`
}

type SearchNearbyResponse struct {
	Drivers  []*DriverLocation `json:"drivers"`
	RadiusKm float32           `json:"radius_km"`
}
