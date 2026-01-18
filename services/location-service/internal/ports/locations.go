package ports

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
)

type LocationRepository interface {
	Save(ctx context.Context, loc *domain.UserLocation) error
	Get(ctx context.Context, userID string) (*domain.UserLocation, error)
	RemoveUserLocation(ctx context.Context, userID string) error
	SearchNearby(ctx context.Context, coordinates domain.Coordinates, radiusKm float32) ([]*domain.UserLocation, error)
}
