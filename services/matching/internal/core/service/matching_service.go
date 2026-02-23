package service

import (
	"context"

	"github.com/google/uuid"
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
	Logger         pkgPorts.Logger
	Metrics        pkgPorts.Metrics
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
	if o.Logger == nil {
		return errors.NewValidationErrorf("logger is required")
	}
	if o.Metrics == nil {
		return errors.NewValidationErrorf("metrics is required")
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
		logger:         opts.Logger,
		metrics:        opts.Metrics,
		contextManager: opts.ContextManager,
	}, nil
}

func (s *MatchingService) MatchRiderToDriver(ctx context.Context, headers map[string]string, request *contracts.RideRequestedEvent) (uuid.UUID, error) {
	// info, ok := s.contextManager.Extract(ctx)
	// if !ok {
	// 	return uuid.Nil, errors.NewPermanentErrorf("missing request info in context")
	// }

	md := metadata.New(headers)
	md.Append("api-key", s.config.LocationService.APIKey)
	ctx = metadata.NewOutgoingContext(ctx, md)
	s.logger.Debug("MatchingService: Getting candidates for ride request", "ride_id", request.RideID, "pickup_location", request.PickupLocation)
	candidates, err := s.client.GetCandidates(ctx, request.PickupLocation)
	if err != nil {
		s.logger.Error("Failed to get candidates from location service", "error", err)
		s.metrics.DependencyFailure("LocationService", "GetCandidates", err.Error())
		return uuid.Nil, err
	}

	if len(candidates) == 0 {
		s.logger.Warn("No drivers available for ride.", "ride_id", request.RideID)
		return uuid.Nil, nil // No candidates found, return nil UUID and no error
	}

	candidate := candidates[0].GetUserId()
	driverID := uuid.MustParse(candidate)
	event := &contracts.RideMatchedEvent{
		RideID:   request.RideID,
		DriverID: driverID,
	}
	// TODO: handle candidates not found
	// For the moment, just pick the first candidate. Implement better matching logic here.
	message := contracts.NewEventMessage(event)
	message.AddHeaders(headers)
	err = s.publisher.Publish(ctx, contracts.TopicMatching, message)
	s.logger.Debug("Published RideMatchedEvent", "ride_id", request.RideID, "driver_id", driverID)
	if err != nil {
		s.logger.Error("Failed to publish RideMatchedEvent", "error", err)
		s.metrics.DependencyFailure("EventPublisher", "Publish", err.Error())
		return uuid.Nil, err
	}

	// Implement matching logic here
	return driverID, nil
}
