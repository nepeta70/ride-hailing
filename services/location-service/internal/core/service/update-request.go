package service

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
)

type UpdateRequest struct {
	UserID      string
	UserType    enums.UserType
	Coordinates domain.Coordinates
	Accuracy    float32
	Heading     float32
	Speed       float32
	CapturedAt  time.Time
}

type SearchNearbyRequest struct {
	UserID      string
	UserType    enums.UserType
	Coordinates domain.Coordinates
	RadiusKm    float32
}

// Validate ensures the GPS data is physically valid
func (r UpdateRequest) Validate() error {
	if r.UserID == "" {
		return errors.NewBusinessError("user_id is required")
	}

	// Latitude range: [-90, 90]
	if r.Coordinates.Latitude < -90 || r.Coordinates.Latitude > 90 {
		return errors.NewBusinessErrorf("invalid latitude: %f", r.Coordinates.Latitude)
	}

	// Longitude range: [-180, 180]
	if r.Coordinates.Longitude < -180 || r.Coordinates.Longitude > 180 {
		return errors.NewBusinessErrorf("invalid longitude: %f", r.Coordinates.Longitude)
	}

	// Accuracy should be positive (meters)
	if r.Accuracy < 0 {
		return errors.NewBusinessError("accuracy cannot be negative")
	}

	// Heading range: [0, 360]
	if r.Heading < 0 || r.Heading > 360 {
		return errors.NewBusinessError("heading must be between 0 and 360 degrees")
	}

	return nil
}
