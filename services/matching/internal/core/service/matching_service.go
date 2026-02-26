package service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/matching/internal/config"
	"github.com/nepeta70/ride-hailing/services/matching/internal/ports"
	"google.golang.org/grpc/metadata"
)

type MatchingServiceOpts struct {
	Config         *config.Config
	Client         ports.GetCandidates
	Publisher      pkgPorts.EventPublisher
	ContextManager *ctxmgr.ContextManager
	Telemetry      pkgPorts.TelemetryProvider
}

func (o *MatchingServiceOpts) Validate() error {
	if o.Config == nil {
		return errors.NewValidationErrorf("config is required")
	}
	if o.Client == nil {
		return errors.NewValidationErrorf("client is required")
	}
	if o.Publisher == nil {
		return errors.NewValidationErrorf("publisher is required")
	}
	if o.Telemetry == nil {
		return errors.NewValidationErrorf("telemetry is required")
	}
	if o.ContextManager == nil {
		return errors.NewValidationErrorf("context manager is required")
	}
	return nil
}

type MatchingService struct {
	config         *config.Config
	client         ports.GetCandidates
	publisher      pkgPorts.EventPublisher
	logger         pkgPorts.Logger
	metrics        pkgPorts.Metrics
	contextManager *ctxmgr.ContextManager
}

func NewMatchingService(opts *MatchingServiceOpts) (*MatchingService, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &MatchingService{
		config:         opts.Config,
		client:         opts.Client,
		publisher:      opts.Publisher,
		logger:         opts.Telemetry.Logger(),
		metrics:        opts.Telemetry.Metrics(),
		contextManager: opts.ContextManager,
	}, nil
}

func (s *MatchingService) MatchRiderToDriver(ctx context.Context, headers map[string]string, request *contracts.RideRequestedEvent) (uuid.UUID, error) {
	md := metadata.New(headers)
	md.Append("api-key", s.config.LocationService.APIKey)
	ctx = metadata.NewOutgoingContext(ctx, md)
	s.logger.DebugContext(ctx, "MatchingService: Getting candidates for ride request", "ride_id", request.RideID.String(), "pickup_location", request.PickupLocation)
	candidates, err := s.client.GetCandidates(ctx, request.PickupLocation)
	if err != nil {
		return uuid.Nil, err
	}

	if len(candidates) == 0 {
		s.logger.WarnContext(ctx, "MatchingService: No drivers available for ride.", "ride_id", request.RideID.String())
		return uuid.Nil, nil // No candidates found, return nil UUID and no error
	}

	slices.SortFunc(candidates, func(a, b *locationv1.SearchNearbyDriversResponse_Driver) int {
		xa := s.SortWeight(a)
		xb := s.SortWeight(b)

		if xa < xb {
			return -1
		}
		if xa > xb {
			return 1
		}

		return 0
	})
	candidate := candidates[0].GetUserId()
	driverID := uuid.MustParse(candidate)
	event := &contracts.RideMatchedEvent{
		RideID:   request.RideID,
		DriverID: driverID,
	}

	message := contracts.NewEventMessage(event)
	message.AddHeaders(headers)
	err = s.publisher.Publish(ctx, contracts.TopicMatching, message)
	s.logger.DebugContext(ctx, "MatchingService: Published RideMatchedEvent", "ride_id", request.RideID.String(), "driver_id", driverID.String())
	if err != nil {
		s.logger.ErrorContext(ctx, "MatchingService: Failed to publish RideMatchedEvent", "error", err)
		s.metrics.DependencyFailure("EventPublisher", "Publish", err.Error())
		return uuid.Nil, err
	}

	// Implement matching logic here
	return driverID, nil
}

func (s *MatchingService) SortWeight(driver *locationv1.SearchNearbyDriversResponse_Driver) float32 {
	return driver.DistanceKm*s.config.Logic.DistanceWeight + float32(driver.AvailableSince.Seconds)*s.config.Logic.AvailabilityWeight
}
