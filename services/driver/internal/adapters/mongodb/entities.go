package mongodb

import (
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/driver/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DriverDoc is the internal representation in MongoDB
type DriverDoc struct {
	ID            bson.ObjectID `bson:"_id,omitempty"`
	UserID        string        `bson:"user_id"` // Indexed
	LicenseNumber string        `bson:"license_number"`
	LicenseExpiry time.Time     `bson:"license_expiry"`
	Vehicle       VehicleDoc    `bson:"vehicle"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
}

func (d *DriverDoc) ToDomain() *domain.Driver {
	return &domain.Driver{
		UserID:        uuid.MustParse(d.UserID),
		LicenseNumber: d.LicenseNumber,
		LicenseExpiry: d.LicenseExpiry,
		Vehicle: &domain.Vehicle{
			Make:              d.Vehicle.Make,
			Model:             d.Vehicle.Model,
			Color:             d.Vehicle.Color,
			LicensePlate:      d.Vehicle.LicensePlate,
			Seats:             d.Vehicle.Seats,
			Category:          d.Vehicle.Category,
			AcceptsPets:       d.Vehicle.AcceptsPets,
			AcceptsWheelchair: d.Vehicle.AcceptsWheelchair,
			AdditionalInfo:    d.Vehicle.AdditionalInfo,
		},
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// FromDomain converts a Domain entity to a MongoDB document
func FromDomain(d *domain.Driver) *DriverDoc {
	return &DriverDoc{
		UserID:        d.UserID.String(),
		LicenseNumber: d.LicenseNumber,
		LicenseExpiry: d.LicenseExpiry,
		Vehicle: VehicleDoc{
			Make:              d.Vehicle.Make,
			Model:             d.Vehicle.Model,
			Color:             d.Vehicle.Color,
			LicensePlate:      d.Vehicle.LicensePlate,
			Seats:             d.Vehicle.Seats,
			Category:          d.Vehicle.Category,
			AcceptsPets:       d.Vehicle.AcceptsPets,
			AcceptsWheelchair: d.Vehicle.AcceptsWheelchair,
			AdditionalInfo:    d.Vehicle.AdditionalInfo,
		},
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

type VehicleDoc struct {
	Make              string `bson:"make"`
	Model             string `bson:"model"`
	Color             string `bson:"color"`
	LicensePlate      string `bson:"license_plate"`
	Seats             int32  `bson:"seats"`
	Category          string `bson:"category"`
	AcceptsPets       bool   `bson:"accepts_pets"`
	AcceptsWheelchair bool   `bson:"accepts_wheelchair"`
	AdditionalInfo    string `bson:"additional_info"`
}
