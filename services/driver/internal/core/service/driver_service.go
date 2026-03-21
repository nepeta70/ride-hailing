package service

import (
	"context"

	"github.com/google/uuid"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/driver/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/driver/internal/ports"
)

type DriverService struct {
	repo      ports.DriverRepository
	telemetry pkgPorts.TelemetryProvider
}

func NewDriverService(repo ports.DriverRepository, telemetry pkgPorts.TelemetryProvider) *DriverService {
	return &DriverService{
		repo:      repo,
		telemetry: telemetry,
	}
}

func (s *DriverService) CreateDriver(ctx context.Context, driver *domain.Driver) (*domain.Driver, error) {
	return nil, nil
}

func (s *DriverService) UpdateDriver(ctx context.Context, driver *domain.Driver) (*domain.Driver, error) {
	return nil, nil
}

func (s *DriverService) GetDriver(ctx context.Context, userID uuid.UUID) (*domain.Driver, error) {
	return nil, nil
}
