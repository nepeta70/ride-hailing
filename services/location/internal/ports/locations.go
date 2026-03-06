package ports

import (
	"context"

	"github.com/google/uuid"
	common "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
)

type LocationRepository interface {
	SaveDriverLocation(ctx context.Context, loc *domain.DriverLocation) error
	SaveDriverStatus(ctx context.Context, status *domain.DriverStatusUpdate) error
	GetDriverLocationAndStatus(ctx context.Context, userID uuid.UUID) (*domain.DriverLocation, error)
	RemoveUserLocation(ctx context.Context, userID uuid.UUID) error
	SearchNearby(ctx context.Context, coordinates *common.Coordinates, radiusKm float32) ([]*domain.DriverLocation, error)
}
