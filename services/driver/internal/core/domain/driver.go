package domain

import (
	"time"

	"github.com/google/uuid"
)

// Driver represents the core business entity for a driver.
type Driver struct {
	UserID        uuid.UUID
	LicenseNumber string
	LicenseExpiry time.Time
	Vehicle       Vehicle
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Vehicle represents the driver's vehicle details within the domain.
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
