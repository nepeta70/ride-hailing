package domain

import "github.com/google/uuid"

type RideStatus string

const (
	RideStatusRequested RideStatus = "REQUESTED"
	RideStatusAccepted  RideStatus = "ACCEPTED"
	RideStatusCancelled RideStatus = "CANCELLED"
	RideStatusOngoing   RideStatus = "ONGOING"
	RideStatusCompleted RideStatus = "COMPLETED"
)

type Ride struct {
	RequestID       uuid.UUID
	ID              uuid.UUID
	RiderID         uuid.UUID
	DriverID        uuid.UUID
	PickupLocation  string
	DropoffLocation string
	Status          RideStatus
}
