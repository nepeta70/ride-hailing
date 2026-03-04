package service

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ctxmgr"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/matching/internal/config"
	"github.com/nepeta70/ride-hailing/services/matching/internal/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	telemetry      pkgPorts.TelemetryProvider
	contextManager *ctxmgr.ContextManager
	semaphore      chan struct{}
	mu             sync.RWMutex
	activeMatches  map[uuid.UUID]context.CancelFunc
}

func NewMatchingService(opts *MatchingServiceOpts) (*MatchingService, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &MatchingService{
		config:         opts.Config,
		client:         opts.Client,
		publisher:      opts.Publisher,
		telemetry:      opts.Telemetry,
		contextManager: opts.ContextManager,
		semaphore:      make(chan struct{}, 10000),
		activeMatches:  make(map[uuid.UUID]context.CancelFunc),
	}, nil
}

func (s *MatchingService) MatchRide(ctx context.Context, headers map[string]string, request *contracts.RideRequestedEvent) {
	ctx = s.telemetry.Propagator().Extract(ctx, propagation.MapCarrier(headers))
	ctx, parentSpan := s.telemetry.Tracer().Start(ctx, "Matching Service: MatchRiderToDriver",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("ride_id", request.RideID.String()),
			attribute.String("pickup_location", request.PickupLocation.String()),
			attribute.String("dropoff_location", request.DropoffLocation.String()),
		),
	)

	defer parentSpan.End()
	bgCtx := context.WithoutCancel(ctx)

	go func(ctx context.Context) {
		s.semaphore <- struct{}{}
		defer func() { <-s.semaphore }()

		bgCtx, span := s.telemetry.Tracer().Start(bgCtx, "MatchingService.BackgroundMatchingLoop",
			trace.WithAttributes(attribute.String("ride.id", request.RideID.String())))
		defer span.End()

		timeoutCtx, timeoutCancel := context.WithTimeout(bgCtx, s.config.Logic.MatchingTimeout)
		matchCtx, manualCancel := context.WithCancel(timeoutCtx)

		defer timeoutCancel()
		defer manualCancel()

		// 5. Register the "Stop Button" in our Map
		s.mu.Lock()
		s.activeMatches[request.RideID] = manualCancel
		s.mu.Unlock()

		// Cleanup the map when the goroutine finishes for any reason
		defer func() {
			s.mu.Lock()
			delete(s.activeMatches, request.RideID)
			s.mu.Unlock()
		}()

		ticker := time.NewTicker(s.config.Logic.MatchingRetryInterval)
		defer ticker.Stop()

		for {
			select {
			case <-matchCtx.Done():
				// DEADLINE REACHED
				s.publishEvent(bgCtx, &contracts.MatchingTimeoutEvent{RideID: request.RideID}, headers)
				return

			case <-ctx.Done():
				return

			case <-ticker.C:
				// TRY TO FIND CANDIDATES
				md := metadata.New(headers)
				md.Set("api-key", s.config.LocationService.APIKey)
				md.Set("timestamp", time.Now().UTC().Format(time.RFC3339))

				// Wrap the background context with the outgoing metadata
				callCtx := metadata.NewOutgoingContext(bgCtx, md)
				s.telemetry.Logger().DebugContext(callCtx, "MatchingService: Getting candidates for ride request", "ride_id", request.RideID.String(), "pickup_location", request.PickupLocation)
				candidates, err := s.client.GetCandidates(callCtx, request.PickupLocation, headers)
				if err != nil {
					span.RecordError(err)
					s.telemetry.Logger().ErrorContext(callCtx, "MatchingService: Failed to get candidates", "ride_id", request.RideID.String(), "error", err)
					continue
				}

				if len(candidates) == 0 {
					span.AddEvent("No candidates found for ride request", trace.WithAttributes(
						attribute.String("ride.id", request.RideID.String()),
					))
					s.telemetry.Logger().WarnContext(callCtx, "MatchingService: No drivers available for ride.", "ride_id", request.RideID.String())
					continue
				}
				span.AddEvent("Candidates found for ride request", trace.WithAttributes(
					attribute.String("ride.id", request.RideID.String()),
					attribute.Int("candidates.count", len(candidates)),
				))
				s.sortCandidates(candidates)

				driverID, err := s.tryToReserveDriver(callCtx, candidates, headers)
				if err != nil {
					span.RecordError(err)
					s.telemetry.Logger().ErrorContext(callCtx, "MatchingService: Failed to reserve driver", "ride_id", request.RideID.String(), "error", err)
					continue
				}
				event := &contracts.RideMatchedEvent{
					RideID:   request.RideID,
					DriverID: driverID,
				}
				s.publishEvent(callCtx, event, headers)

				span.SetAttributes(attribute.String("driver.id", driverID.String()))
				return
			}
		}
	}(ctx)
}

func (s *MatchingService) HandleCancelRide(ctx context.Context, request *contracts.RideCanceledEvent) {
	s.mu.RLock()
	cancel, exists := s.activeMatches[request.RideID]
	s.mu.RUnlock()

	if exists {
		cancel() // This sends the signal to the case <-matchCtx.Done()
		s.telemetry.Logger().InfoContext(ctx, "Ride cancelled by user", "ride_id", request.RideID)
	}
}

func (s *MatchingService) tryToReserveDriver(ctx context.Context, candidates []*locationv1.SearchNearbyDriversResponse_Driver, headers map[string]string) (uuid.UUID, error) {
	span := trace.SpanFromContext(ctx)

	for _, candidate := range candidates {
		span.AddEvent("Attempting to reserve driver", trace.WithAttributes(
			attribute.String("driver.id", candidate.GetUserId()),
		))
		err := s.client.UpdateDriverStatus(ctx, uuid.MustParse(candidate.GetUserId()), contracts.DriverStatusReserved, headers)
		if err == nil {
			span.AddEvent("Successfully reserved driver", trace.WithAttributes(
				attribute.String("driver.id", candidate.GetUserId()),
			))
			return uuid.MustParse(candidate.GetUserId()), nil

		} else {
			span.RecordError(err)
			s.telemetry.Logger().ErrorContext(ctx, "MatchingService: Failed to update driver status", "driver_id", candidate.GetUserId(), "error", err)
		}
	}

	return uuid.Nil, nil // No candidates could be reserved, return nil UUID and no error
}

func (s *MatchingService) sortCandidates(candidates []*locationv1.SearchNearbyDriversResponse_Driver) error {
	slices.SortFunc(candidates, func(a, b *locationv1.SearchNearbyDriversResponse_Driver) int {
		xa := s.sortWeight(a)
		xb := s.sortWeight(b)

		if xa < xb {
			return -1
		}
		if xa > xb {
			return 1
		}

		return 0
	})

	return nil
}

func (s *MatchingService) sortWeight(driver *locationv1.SearchNearbyDriversResponse_Driver) float32 {
	return driver.DistanceKm*s.config.Logic.DistanceWeight + float32(driver.AvailableSince.Seconds)*s.config.Logic.AvailabilityWeight
}

func (s *MatchingService) publishEvent(ctx context.Context, event contracts.Event, headers map[string]string) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("Publishing Event", trace.WithAttributes(
		attribute.String("event.type", event.EventType().String()),
	))

	message := contracts.NewEventMessage(event)
	s.telemetry.Logger().DebugContext(ctx, "MatchingService: Publishing Event", "event_type", message.EventType.String(), "payload", message.Payload, "headers", message.Headers)
	message.AddHeaders(headers)
	err := s.publisher.Publish(ctx, contracts.TopicMatching, message)
	s.telemetry.Logger().DebugContext(ctx, "MatchingService: Published Event", "event_type", message.EventType.String())
	if err != nil {
		span.RecordError(err)
		s.telemetry.Logger().ErrorContext(ctx, "MatchingService: Failed to publish Event", "error", err)
		s.telemetry.Metrics().DependencyFailure("EventPublisher", "Publish", err.Error())
		return
	}
	span.SetStatus(codes.Ok, "Event published successfully")
}
