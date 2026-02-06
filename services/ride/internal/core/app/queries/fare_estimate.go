package queries

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type FareEstimateRequest struct {
	RequestId       string
	PickupLocation  string
	DropoffLocation string
	CountryCode     string
}

func (q *FareEstimateRequest) Validate() error {
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
	fareRateRepo        ports.FareRatesReadRepository
}

func NewFareEstimateHandler(config *config.Config, storage ports.StorageBundle, directionsEstimator *service.DirectionsEstimator, directionsService ports.DirectionsService) *FareEstimateHandler {
	return &FareEstimateHandler{
		config:              config,
		countryCache:        storage.CountryCache(),
		directionsEstimator: directionsEstimator,
		directionsService:   directionsService,
		fareRateRepo:        storage.FareRatesReadRepo(),
	}
}

func (h *FareEstimateHandler) Handle(ctx context.Context, query FareEstimateRequest) (*domain.Fares, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	err := query.Validate()
	if err != nil {
		return nil, err
	}

	country, exists := h.countryCache.GetCountryByCode(ctx, query.CountryCode)
	if !exists {
		return nil, errors.NewErrNotFound("country not found")
	}

	fareRates, err := h.fareRateRepo.GetRatesByCountry(ctx, query.CountryCode)
	if err != nil {
		return nil, err
	}

	n := len(fareRates)
	if n == 0 {
		return nil, errors.NewErrNotFound("no fare rates found for country")
	}

	directions, err := h.directionsService.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)
	if err != nil || directions == nil {
		directions, err = h.directionsEstimator.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)

		if err != nil {
			return nil, err
		}
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

	return &domain.Fares{
		ID:                       uuid.New(),
		EstimatedDistanceKm:      distanceInKm,
		EstimatedDurationMinutes: directions.DurationMinutes,
		ETA:                      directions.ArrivalTime,
		Currency:                 country.Currency,
		Fares:                    fares,
	}, nil
}
