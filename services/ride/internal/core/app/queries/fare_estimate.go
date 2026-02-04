package queries

import (
	"context"

	"github.com/docker/distribution/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type FareEstimateRequest struct {
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
	directionsEstimator *service.DirectionsEstimator
	directionsService   ports.DirectionsService
	fareRateRepo        ports.FareRatesReadRepository
}

func NewFareEstimateHandler(config *config.Config, directionsEstimator *service.DirectionsEstimator, directionsService ports.DirectionsService, fareRateRepo ports.FareRatesReadRepository) *FareEstimateHandler {
	return &FareEstimateHandler{
		config:              config,
		directionsEstimator: directionsEstimator,
		directionsService:   directionsService,
		fareRateRepo:        fareRateRepo,
	}
}

func (h *FareEstimateHandler) Handle(ctx context.Context, query FareEstimateRequest) (*domain.Fare, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	err := query.Validate()
	if err != nil {
		return nil, err
	}

	directions, err := h.directionsService.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)
	if err != nil || directions == nil {
		directions, err = h.directionsEstimator.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)

		if err != nil {
			return nil, err
		}
	}

	distanceInKm := directions.Distance / 1000.0
	return &domain.Fare{
		ID:                  uuid.Generate(),
		EstimatedDistanceKm: distanceInKm,
		EstimatedDuration:   directions.Duration,
		ETA:                 directions.ArrivalTime,
		Fare:                h.config.FareConfig.BaseFare + h.config.FareConfig.FarePerMinute*directions.Duration.Minutes() + h.config.FareConfig.FarePerKm*distanceInKm,
		Currency:            h.config.FareConfig.DefaultCurrency,
	}, nil
}
