package domain

import (
	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/actor/grain"
	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/shopspring/decimal"
)

type RideStatus string

const (
	RideStatusNew       RideStatus = "NEW"
	RideStatusRequested RideStatus = "REQUESTED"
	RideStatusMatched   RideStatus = "MATCHED"
	RideStatusAccepted  RideStatus = "ACCEPTED"
	RideStatusCancelled RideStatus = "CANCELLED"
	RideStatusStarted   RideStatus = "STARTED"
	RideStatusCompleted RideStatus = "COMPLETED"
	RideStatusTimedOut  RideStatus = "TIMED_OUT"
)

func (s RideStatus) IsValid() bool {
	switch s {
	case RideStatusNew, RideStatusRequested, RideStatusMatched, RideStatusAccepted, RideStatusCancelled, RideStatusStarted, RideStatusCompleted:
		return true
	default:
		return false
	}
}

func (s RideStatus) String() string {
	return string(s)
}

const (
	RideGrainKind enums.AggregateType = "Ride"
)

// RideCore represents the immutable "contract" of the ride
type RideCore struct {
	RequestID       uuid.UUID         `json:"request_id"`
	RiderID         uuid.UUID         `json:"rider_id"`
	PickupLocation  *core.Coordinates `json:"pickup_location"`
	DropoffLocation *core.Coordinates `json:"dropoff_location"`
	ServiceType     string            `json:"service_type"`
	Fare            decimal.Decimal   `json:"fare"`
	Currency        string            `json:"currency"`
}

// RideStatusState represents the mutable lifecycle of the ride
type RideState struct {
	DriverID *uuid.UUID `json:"driver_id"`
	Status   RideStatus `json:"status"`
}

type GrainData struct {
	Identity *grain.GrainIdentity   `json:"identity"`
	Core     *RideCore              `json:"core"`
	State    *RideState             `json:"state"`
	Message  ports.MessageInterface `json:"message"`
	Version  int                    `json:"version"`
}
