package ports

import (
	"context"

	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
)

type DirectionsService interface {
	GetDirections(ctx context.Context, origin, destination *core.Coordinates) (*domain.DirectionsResponse, error)
}
