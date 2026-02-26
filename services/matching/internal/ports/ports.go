package ports

import (
	"context"

	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	domain "github.com/nepeta70/ride-hailing/internal/pkg/core"
)

type GetCandidates interface {
	GetCandidates(ctx context.Context, coords *domain.Coordinates) ([]*locationv1.SearchNearbyDriversResponse_Driver, error)
}
