package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
)

type UpdateRequest struct {
	UserID      uuid.UUID
	UserType    enums.UserType
	Coordinates domain.Coordinates
	Accuracy    float32
	Heading     float32
	Speed       float32
	CapturedAt  time.Time
	Status      contracts.DriverStatus
}

type SearchNearbyRequest struct {
	UserID      uuid.UUID
	UserType    enums.UserType
	Coordinates domain.Coordinates
	RadiusKm    float32
}

// Validate ensures the GPS data is physically valid
func (r UpdateRequest) Validate() error {
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

	// Speed should be non-negative (m/s)
	if r.Speed < 0 {
		return errors.NewBusinessError("speed cannot be negative")
	}

	// Status should be one of the defined constants

	return nil
}
