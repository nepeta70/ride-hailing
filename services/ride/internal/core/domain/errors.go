package domain

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

var (
	// ErrRideNotFound indicates that the requested ride does not exist
	ErrRideNotFound  = errors.NewBusinessError("ride not found")
	ErrDuplicateRide = errors.NewBusinessError("ride already exists")
)
