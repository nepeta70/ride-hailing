package ports

import (
	"context"

	"github.com/google/uuid"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	domain "github.com/nepeta70/ride-hailing/internal/pkg/core"
)

type GetCandidates interface {
	GetCandidates(ctx context.Context, coords *domain.Coordinates, radiusKm float32, headers map[string]string) (*locationv1.SearchNearbyDriversResponse, error)
	UpdateDriverStatus(ctx context.Context, driverID uuid.UUID, status contracts.DriverStatus, headers map[string]string) error
}
