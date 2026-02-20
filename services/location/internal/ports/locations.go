package ports

import (
	"context"

	"github.com/google/uuid"
	common "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
)

type LocationRepository interface {
	Save(ctx context.Context, loc *domain.DriverLocation) error
	Get(ctx context.Context, userID uuid.UUID) (*domain.DriverLocation, error)
	RemoveUserLocation(ctx context.Context, userID uuid.UUID) error
	SearchNearby(ctx context.Context, coordinates *common.Coordinates, radiusKm float32) ([]*domain.DriverLocation, error)
}
