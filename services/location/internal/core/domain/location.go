package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
)

type DriverLocation struct {
	UserID      uuid.UUID              `json:"user_id"`
	UserType    enums.UserType         `json:"user_type"`
	Coordinates Coordinates            `json:"coordinates"`
	Accuracy    float32                `json:"accuracy"`
	Heading     float32                `json:"heading"`
	Speed       float32                `json:"speed"`
	CapturedAt  time.Time              `json:"captured_at"`
	Status      contracts.DriverStatus `json:"status"`
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
