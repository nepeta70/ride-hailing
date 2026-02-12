package domain

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/shopspring/decimal"
)

type RideStatus string

const (
	RideStatusNew       RideStatus = "NEW"
	RideStatusRequested RideStatus = "REQUESTED"
	RideStatusAccepted  RideStatus = "ACCEPTED"
	RideStatusCancelled RideStatus = "CANCELLED"
	RideStatusOngoing   RideStatus = "ONGOING"
	RideStatusCompleted RideStatus = "COMPLETED"
)

const (
	RideGrainKind enums.AggregateType = "Ride"
)

// RideCore represents the immutable "contract" of the ride
type RideCore struct {
	RequestID       uuid.UUID       `json:"request_id"`
	RiderID         uuid.UUID       `json:"rider_id"`
	PickupLocation  string          `json:"pickup_location"`
	DropoffLocation string          `json:"dropoff_location"`
	ServiceType     string          `json:"service_type"`
	Fare            decimal.Decimal `json:"fare"`
	Currency        string          `json:"currency"`
}

// RideStatusState represents the mutable lifecycle of the ride
type RideState struct {
	DriverID *uuid.UUID `json:"driver_id"`
	Status   RideStatus `json:"status"`
}

type GrainData struct {
	Identity *grain.GrainIdentity `json:"identity"`
	Core     *RideCore            `json:"core"`
	State    *RideState           `json:"state"`
	Command  ports.Command        `json:"command"`
	Version  int                  `json:"version"`
}
