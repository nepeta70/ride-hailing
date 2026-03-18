package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	common "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	pkgErrors "github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgMocks "github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	"github.com/nepeta70/ride-hailing/services/location/internal/adapters/mocks"
	"github.com/nepeta70/ride-hailing/services/location/internal/config"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
	. "github.com/nepeta70/ride-hailing/services/location/internal/core/service"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestNewLocationService(t *testing.T) {
	tests := []struct {
		name    string
		opts    *LocationServiceOpts
		wantErr bool
	}{
		{
			name: "successful creation with valid opts",
			opts: &LocationServiceOpts{
				Config:    config.DefaultConfig(),
				Repo:      &mocks.MockLocationRepository{},
				Telemetry: &pkgMocks.MockTelemetryProvider{},
			},
			wantErr: false,
		},
		{
			name: "panic when config is nil",
			opts: &LocationServiceOpts{
				Config:    nil,
				Repo:      &mocks.MockLocationRepository{},
				Telemetry: &pkgMocks.MockTelemetryProvider{},
			},
			wantErr: true,
		},
		{
			name: "panic when repo is nil",
			opts: &LocationServiceOpts{
				Config:    config.DefaultConfig(),
				Repo:      nil,
				Telemetry: &pkgMocks.MockTelemetryProvider{},
			},
			wantErr: true,
		},
		{
			name: "panic when telemetry is nil",
			opts: &LocationServiceOpts{
				Config:    config.DefaultConfig(),
				Repo:      &mocks.MockLocationRepository{},
				Telemetry: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Panics(t, func() {
					NewLocationService(tt.opts)
				})
			} else {
				service := NewLocationService(tt.opts)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestLocationService_Update(t *testing.T) {
	validDriverID := uuid.New()
	validCoordinates := common.Coordinates{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		req       *UpdateDriverLocationRequest
		mockSetup func(*mocks.MockLocationRepository)
		wantErr   bool
	}{
		{
			name: "successful location update",
			ctx:  context.Background(),
			req: &UpdateDriverLocationRequest{
				DriverID:        validDriverID,
				SenderType:      enums.SenderTypeDriver,
				Coordinates:     validCoordinates,
				Accuracy:        10.5,
				Heading:         45.0,
				Speed:           15.5,
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("SaveDriverLocation", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "failed location update",
			ctx:  context.Background(),
			req: &UpdateDriverLocationRequest{
				DriverID:        validDriverID,
				SenderType:      enums.SenderTypeDriver,
				Coordinates:     validCoordinates,
				Accuracy:        10.5,
				Heading:         45.0,
				Speed:           15.5,
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("SaveDriverLocation", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "invalid coordinates",
			ctx:  context.Background(),
			req: &UpdateDriverLocationRequest{
				DriverID:   validDriverID,
				SenderType: enums.SenderTypeDriver,
				Coordinates: common.Coordinates{
					Latitude:  200.0, // Invalid: exceeds max latitude
					Longitude: -74.0060,
				},
				Accuracy:        10.5,
				Heading:         45.0,
				Speed:           15.5,
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
		{
			name: "negative accuracy",
			ctx:  context.Background(),
			req: &UpdateDriverLocationRequest{
				DriverID:        validDriverID,
				SenderType:      enums.SenderTypeDriver,
				Coordinates:     validCoordinates,
				Accuracy:        -5.0, // Invalid: negative
				Heading:         45.0,
				Speed:           15.5,
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
		{
			name: "invalid heading",
			ctx:  context.Background(),
			req: &UpdateDriverLocationRequest{
				DriverID:        validDriverID,
				SenderType:      enums.SenderTypeDriver,
				Coordinates:     validCoordinates,
				Accuracy:        10.5,
				Heading:         450.0, // Invalid: exceeds 360
				Speed:           15.5,
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
		{
			name: "negative speed",
			ctx:  context.Background(),
			req: &UpdateDriverLocationRequest{
				DriverID:        validDriverID,
				SenderType:      enums.SenderTypeDriver,
				Coordinates:     validCoordinates,
				Accuracy:        10.5,
				Heading:         45.0,
				Speed:           -10.0, // Invalid: negative
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
		{
			name: "nil driver ID",
			ctx:  context.Background(),
			req: &UpdateDriverLocationRequest{
				DriverID:        uuid.Nil, // Invalid
				SenderType:      enums.SenderTypeDriver,
				Coordinates:     validCoordinates,
				Accuracy:        10.5,
				Heading:         45.0,
				Speed:           15.5,
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
		{
			name: "context cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			req: &UpdateDriverLocationRequest{
				DriverID:        validDriverID,
				SenderType:      enums.SenderTypeDriver,
				Coordinates:     validCoordinates,
				Accuracy:        10.5,
				Heading:         45.0,
				Speed:           15.5,
				Status:          contracts.DriverStatusAvailable,
				CapturedAt:      time.Now(),
				StatusUpdatedAt: time.Now(),
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockLocationRepository{}
			tt.mockSetup(mockRepo)

			mockTelemetry := &pkgMocks.MockTelemetryProvider{}
			mockTelemetry.On("Tracer").Return(&pkgMocks.MockTracer{})

			service := NewLocationService(&LocationServiceOpts{
				Config:    config.DefaultConfig(),
				Repo:      mockRepo,
				Telemetry: mockTelemetry,
			})

			err := service.Update(tt.ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertCalled(t, "SaveDriverLocation", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestLocationService_Get(t *testing.T) {
	driverID := uuid.New()

	tests := []struct {
		name      string
		ctx       context.Context
		userID    uuid.UUID
		mockSetup func(*mocks.MockLocationRepository)
		wantErr   bool
		validate  func(*domain.DriverLocation)
	}{
		{
			name:   "successful get location",
			ctx:    context.Background(),
			userID: driverID,
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("GetDriverLocationAndStatus", mock.Anything, driverID).Return(
					&domain.DriverLocation{
						UserID: driverID,
						Coordinates: common.Coordinates{
							Latitude:  40.7128,
							Longitude: -74.0060,
						},
					},
					nil,
				)
			},
			wantErr: false,
			validate: func(loc *domain.DriverLocation) {
				assert.NotNil(t, loc)
				assert.Equal(t, driverID, loc.UserID)
			},
		},
		{
			name:   "location not found",
			ctx:    context.Background(),
			userID: driverID,
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("GetDriverLocationAndStatus", mock.Anything, driverID).Return(
					nil,
					pkgErrors.NewErrNotFoundf("location not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "context cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			userID:    driverID,
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockLocationRepository{}
			tt.mockSetup(mockRepo)

			mockTelemetry := &pkgMocks.MockTelemetryProvider{}

			service := NewLocationService(&LocationServiceOpts{
				Config:    config.DefaultConfig(),
				Repo:      mockRepo,
				Telemetry: mockTelemetry,
			})

			loc, err := service.Get(tt.ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(loc)
				}
			}
		})
	}
}

func TestLocationService_RemoveUserLocation(t *testing.T) {
	driverID := uuid.New()

	tests := []struct {
		name      string
		ctx       context.Context
		userID    uuid.UUID
		mockSetup func(*mocks.MockLocationRepository)
		wantErr   bool
	}{
		{
			name:   "successful removal",
			ctx:    context.Background(),
			userID: driverID,
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("RemoveUserLocation", mock.Anything, driverID).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "removal failed",
			ctx:    context.Background(),
			userID: driverID,
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("RemoveUserLocation", mock.Anything, driverID).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "context cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			userID:    driverID,
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockLocationRepository{}
			tt.mockSetup(mockRepo)

			mockTelemetry := &pkgMocks.MockTelemetryProvider{}

			service := NewLocationService(&LocationServiceOpts{
				Config:    config.DefaultConfig(),
				Repo:      mockRepo,
				Telemetry: mockTelemetry,
			})

			err := service.RemoveUserLocation(tt.ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertCalled(t, "RemoveUserLocation", mock.Anything, driverID)
			}
		})
	}
}

func TestLocationService_SearchNearby(t *testing.T) {
	driverID := uuid.New()
	searchCoordinates := common.Coordinates{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	tests := []struct {
		name        string
		ctx         context.Context
		coordinates *common.Coordinates
		mockSetup   func(*mocks.MockLocationRepository)
		wantErr     bool
		validate    func([]*domain.DriverLocation)
	}{
		{
			name:        "successful search with results",
			ctx:         context.Background(),
			coordinates: &searchCoordinates,
			mockSetup: func(m *mocks.MockLocationRepository) {
				drivers := []*domain.DriverLocation{
					{
						UserID: driverID,
						Coordinates: common.Coordinates{
							Latitude:  40.7128,
							Longitude: -74.0060,
						},
					},
				}
				m.On("SearchNearby", mock.Anything, mock.Anything, mock.AnythingOfType("float32")).Return(drivers, nil)
			},
			wantErr: false,
			validate: func(drivers []*domain.DriverLocation) {
				assert.Len(t, drivers, 1)
				assert.Equal(t, driverID, drivers[0].UserID)
			},
		},
		{
			name:        "no drivers found",
			ctx:         context.Background(),
			coordinates: &searchCoordinates,
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("SearchNearby", mock.Anything, mock.Anything, mock.AnythingOfType("float32")).Return(
					nil,
					pkgErrors.NewErrNotFoundf("no drivers found"),
				)
			},
			wantErr: true,
		},
		{
			name: "context cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			coordinates: &searchCoordinates,
			mockSetup:   func(m *mocks.MockLocationRepository) {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockLocationRepository{}
			tt.mockSetup(mockRepo)

			mockTelemetry := pkgMocks.NewMockTelemetryProvider()
			//mockTelemetry.On("Tracer").Return(&pkgMocks.MockTracer{})
			cfg := config.DefaultConfig()
			cfg.Logic.LocationTTLSeconds = 15
			cfg.Init()

			service := NewLocationService(&LocationServiceOpts{
				Config:    cfg,
				Repo:      mockRepo,
				Telemetry: mockTelemetry,
			})

			res, err := service.SearchNearby(tt.ctx, tt.coordinates, 1.0)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(res.Drivers)
				}
			}
		})
	}
}

func TestLocationService_UpdateDriverStatus(t *testing.T) {
	driverID := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		ctx       context.Context
		req       *UpdateDriverStatusRequest
		mockSetup func(*mocks.MockLocationRepository)
		wantErr   bool
	}{
		{
			name: "successful status update",
			ctx:  context.Background(),
			req: &UpdateDriverStatusRequest{
				DriverID:        driverID,
				Status:          contracts.DriverStatusAvailable,
				StatusUpdatedAt: now,
			},
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("SaveDriverStatus", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "failed status update",
			ctx:  context.Background(),
			req: &UpdateDriverStatusRequest{
				DriverID:        driverID,
				Status:          contracts.DriverStatusAvailable,
				StatusUpdatedAt: now,
			},
			mockSetup: func(m *mocks.MockLocationRepository) {
				m.On("SaveDriverStatus", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "nil driver ID",
			ctx:  context.Background(),
			req: &UpdateDriverStatusRequest{
				DriverID:        uuid.Nil,
				Status:          contracts.DriverStatusAvailable,
				StatusUpdatedAt: now,
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
		{
			name: "zero status updated at",
			ctx:  context.Background(),
			req: &UpdateDriverStatusRequest{
				DriverID:        driverID,
				Status:          contracts.DriverStatusAvailable,
				StatusUpdatedAt: time.Time{},
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
		{
			name: "context cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			req: &UpdateDriverStatusRequest{
				DriverID:        driverID,
				Status:          contracts.DriverStatusAvailable,
				StatusUpdatedAt: now,
			},
			mockSetup: func(m *mocks.MockLocationRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockLocationRepository{}
			tt.mockSetup(mockRepo)

			mockTelemetry := &pkgMocks.MockTelemetryProvider{}
			mockTelemetry.On("Tracer").Return(&pkgMocks.MockTracer{})

			service := NewLocationService(&LocationServiceOpts{
				Config:    config.DefaultConfig(),
				Repo:      mockRepo,
				Telemetry: mockTelemetry,
			})

			err := service.UpdateDriverStatus(tt.ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertCalled(t, "SaveDriverStatus", mock.Anything, mock.Anything)
			}
		})
	}
}
