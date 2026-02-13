package ports

import (
	"context"

	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

type DirectionsService interface {
	GetDirections(ctx context.Context, origin, destination string) (*domain.DirectionsResponse, error)
}
