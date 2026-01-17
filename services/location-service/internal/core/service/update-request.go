package service

import (
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type UpdateRequest struct {
	EntityID   string
	Latitude   float64
	Longitude  float64
	Accuracy   float32
	Heading    float32
	Speed      float32
	CapturedAt time.Time
}

// Validate ensures the GPS data is physically valid
func (r UpdateRequest) Validate() error {
	if r.EntityID == "" {
		return errors.NewBusinessError("entity_id is required")
	}

	// Latitude range: [-90, 90]
	if r.Latitude < -90 || r.Latitude > 90 {
		return errors.NewBusinessErrorf("invalid latitude: %f", r.Latitude)
	}

	// Longitude range: [-180, 180]
	if r.Longitude < -180 || r.Longitude > 180 {
		return errors.NewBusinessErrorf("invalid longitude: %f", r.Longitude)
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
