package ports

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/location-service/internal/core/domain"
)

type LocationRepository interface {
	Save(ctx context.Context, loc *domain.Location) error
	Get(ctx context.Context, entityID string) (*domain.Location, error)
}
