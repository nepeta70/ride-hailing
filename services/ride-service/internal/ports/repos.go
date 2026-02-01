package ports

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/domain"
)

type ReadFareRepository interface {
	// Define methods for reading fare data
}

type WriteFareRepository interface {
	// Define methods for writing fare data
}
type WriteRideRepository interface {
	// Define methods for writing ride data
}
type ReadRideRepository interface {
	// Define methods for reading ride data
}

type DirectionsService interface {
	// Define methods for interacting with directions services
	GetDirections(ctx context.Context, origin, destination string) (*domain.DirectionsResponse, error)
}
