package commands

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type EstimateFareCommand struct {
	RequestId       string
	PickupLocation  string
	DropoffLocation string
	CountryCode     string
}

func (q *EstimateFareCommand) Validate() error {
	if q.PickupLocation == "" || q.DropoffLocation == "" {
		return errors.NewValidationErrorf("pickup and dropoff locations are required")
	}
	if q.PickupLocation == q.DropoffLocation {
		return errors.NewValidationErrorf("pickup and dropoff locations cannot be the same")
	}
	return nil
}

type FareEstimateHandler struct {
	config              *config.Config
	countryCache        ports.CountryCacheInterface
	directionsEstimator *service.DirectionsEstimator
	directionsService   ports.DirectionsService
	storage             ports.StorageBundle
	logger              pkgPorts.Logger
}

func NewEstimateFareHandler(config *config.Config, logger pkgPorts.Logger, storage ports.StorageBundle, directionsEstimator *service.DirectionsEstimator, directionsService ports.DirectionsService) *FareEstimateHandler {
	return &FareEstimateHandler{
		config:              config,
		countryCache:        storage.CountryCache(),
		directionsEstimator: directionsEstimator,
		directionsService:   directionsService,
		storage:             storage,
		logger:              logger,
	}
}

func (h *FareEstimateHandler) Handle(ctx context.Context, query EstimateFareCommand) (*domain.Fares, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	err := query.Validate()
	if err != nil {
		return nil, err
	}

	country, exists := h.countryCache.GetCountryByCode(ctx, query.CountryCode)
	if !exists {
		h.logger.Error("country not found: %s", query.CountryCode)
		return nil, errors.NewErrNotFound("country not found")
	}

	fareRates, err := h.storage.FareRatesReadRepo().GetRatesByCountry(ctx, query.CountryCode)
	if err != nil {
		return nil, err
	}

	n := len(fareRates)
	if n == 0 {
		h.logger.Error("no fare rates found for country: %s", query.CountryCode)
		return nil, errors.NewErrNotFound("no fare rates found for country")
	}

	directions, err := h.getDirections(ctx, query)
	if err != nil {
		h.logger.Error("failed to get directions: %v", err)
		return nil, errors.NewPermanentErrorf("failed to get directions: %w", err)
	}

	fares := make([]*domain.Fare, 0, n)
	distanceInKm := directions.DistanceMeters / 1000.0
	for _, rate := range fareRates {
		fareAmount := rate.BaseFare + (rate.FarePerKm * distanceInKm) + (rate.FarePerMinute * directions.DurationMinutes.Minutes())
		fares = append(fares, &domain.Fare{
			ServiceType: rate.ServiceType,
			Fare:        math.Max(rate.MinimumFare, fareAmount),
		})
	}

	record := &domain.Fares{
		ID:                       uuid.New(),
		EstimatedDistanceKm:      distanceInKm,
		EstimatedDurationMinutes: directions.DurationMinutes,
		ETA:                      directions.ArrivalTime,
		Currency:                 country.Currency,
		Fares:                    fares,
	}

	err = h.storage.FareWriteRepo().Save(ctx, record)
	if err != nil {
		h.logger.Error("failed to save fare estimate: %v", err)
		return nil, errors.NewPermanentErrorf("failed to save fare estimate: %w", err)
	}
	return record, nil
}

func (h *FareEstimateHandler) getDirections(ctx context.Context, query EstimateFareCommand) (*domain.DirectionsResponse, error) {
	if h.directionsService != nil {
		directions, err := h.directionsService.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)
		if err == nil && directions != nil {
			return directions, nil
		}
	}
	h.logger.Warn("directions service failed, falling back to estimator")

	return h.directionsEstimator.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)
}
