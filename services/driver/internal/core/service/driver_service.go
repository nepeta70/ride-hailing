package service

import (
	"context"
	"time"

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
	driver.UserID = uuid.New()
	now := time.Now().UTC()
	driver.CreatedAt = now
	driver.UpdatedAt = now
	// TODO check duplicated licenses, plates

	if err := driver.Validate(); err != nil {
		return nil, err
	}
	err := s.repo.AddDriver(ctx, driver)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func (s *DriverService) UpdateDriver(ctx context.Context, driver *domain.Driver) (*domain.Driver, error) {
	now := time.Now().UTC()
	driver.UpdatedAt = now

	if err := driver.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateDriver(ctx, driver); err != nil {
		return nil, err
	}
	return driver, nil
}

func (s *DriverService) GetDriver(ctx context.Context, userID uuid.UUID) (*domain.Driver, error) {
	return s.repo.GetDriver(ctx, userID)
}
