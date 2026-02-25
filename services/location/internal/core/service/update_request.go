package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	domain "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type UpdateDriverStatusRequest struct {
	DriverID        uuid.UUID
	UserType        enums.UserType
	Coordinates     domain.Coordinates
	Accuracy        float32
	Heading         float32
	Speed           float32
	CapturedAt      time.Time
	Status          contracts.DriverStatus
	StatusUpdatedAt time.Time
}

// Validate ensures the GPS data is physically valid
func (r UpdateDriverStatusRequest) Validate() error {
	if err := r.Coordinates.Validate(); err != nil {
		return errors.NewValidationErrorf("invalid coordinates: %w", err)
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

	if r.DriverID == uuid.Nil {
		return errors.NewValidationErrorf("user ID cannot be empty")
	}

	return nil
}
