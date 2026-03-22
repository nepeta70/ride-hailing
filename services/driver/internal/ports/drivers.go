package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/driver/internal/core/domain"
)

type DriverReadRepository interface {
	GetDriver(ctx context.Context, userID uuid.UUID) (*domain.Driver, error)
}

type DriverWriteRepository interface {
	AddDriver(ctx context.Context, driver *domain.Driver) error
	UpdateDriver(ctx context.Context, driver *domain.Driver) error
}

type DriverRepository interface {
	DriverReadRepository
	DriverWriteRepository
}