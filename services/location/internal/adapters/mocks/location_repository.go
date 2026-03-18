package mocks

import (
	"context"

	"github.com/google/uuid"
	common "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockLocationRepository is a mock implementation of the LocationRepository interface
type MockLocationRepository struct {
	mock.Mock
}

func (m *MockLocationRepository) SaveDriverLocation(ctx context.Context, loc *domain.DriverLocation) error {
	args := m.Called(ctx, loc)
	return args.Error(0)
}

func (m *MockLocationRepository) SaveDriverStatus(ctx context.Context, status *domain.DriverStatusUpdate) error {
	args := m.Called(ctx, status)
	return args.Error(0)
}

func (m *MockLocationRepository) GetDriverLocationAndStatus(ctx context.Context, userID uuid.UUID) (*domain.DriverLocation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DriverLocation), args.Error(1)
}

func (m *MockLocationRepository) RemoveUserLocation(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockLocationRepository) SearchNearby(ctx context.Context, coordinates *common.Coordinates, radiusKm float32) ([]*domain.DriverLocation, error) {
	args := m.Called(ctx, coordinates, radiusKm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DriverLocation), args.Error(1)
}
