package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type Driver struct {
	UserID        uuid.UUID
	LicenseNumber string
	LicenseExpiry time.Time
	Vehicle       *Vehicle
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (d *Driver) Validate() error {
	if d.UserID == uuid.Nil {
		return errors.NewValidationErrorf("user_id is required")
	}

	if len(d.LicenseNumber) == 0 {
		return errors.NewValidationErrorf("license_number cannot be empty")
	}

	if d.LicenseExpiry.IsZero() {
		return errors.NewValidationErrorf("license_expiry date is required")
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	licenseUTC := d.LicenseExpiry.UTC()

	if licenseUTC.Before(today) {
		return errors.NewValidationErrorf("license has already expired (expired on %s)", d.LicenseExpiry.Format("2006-01-02"))
	}

	// 4. Delegate to Vehicle Validation
	if err := d.Vehicle.Validate(); err != nil {
		return errors.NewValidationErrorf("invalid vehicle info: %w", err)
	}

	return nil
}

func NewDriver(driverID uuid.UUID) *Driver {
	now := time.Now().UTC()
	return &Driver{
		UserID:    driverID,
		CreatedAt: now,
		UpdatedAt: now,
		Vehicle:   &Vehicle{},
	}
}

func UpdateDriver(driverID uuid.UUID) *Driver {
	now := time.Now().UTC()
	return &Driver{
		UserID:    driverID,
		UpdatedAt: now,
		Vehicle:   &Vehicle{},
	}
}

type Vehicle struct {
	Make              string
	Model             string
	Color             string
	LicensePlate      string
	Seats             int32
	Category          string
	AcceptsPets       bool
	AcceptsWheelchair bool
	AdditionalInfo    string
}

func (v *Vehicle) Validate() error {
	if v.Make == "" || v.Model == "" {
		return errors.NewValidationErrorf("vehicle make and model are required")
	}
	if v.LicensePlate == "" {
		return errors.NewValidationErrorf("vehicle license_plate is required")
	}
	if v.Seats < 1 {
		return errors.NewValidationErrorf("vehicle must have at least 1 seat")
	}
	return nil
}
