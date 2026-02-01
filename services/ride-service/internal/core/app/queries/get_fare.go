package queries

import (
	"context"

	"github.com/docker/distribution/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/core/service"
	"github.com/nepeta70/ride-hailing/services/ride-service/internal/ports"
)

type GetFare struct {
	PickupLocation  string
	DropoffLocation string
}

func (q *GetFare) Validate() error {
	if q.PickupLocation == "" || q.DropoffLocation == "" {
		return errors.NewValidationErrorf("pickup and dropoff locations are required")
	}

	return nil
}

type GetFareHandler struct {
	distanceCalculator *service.DirectionsEstimator
	directionsService  ports.DirectionsService
}

func NewGetFareHandler(distanceCalculator *service.DirectionsEstimator, directionsService ports.DirectionsService) GetFareHandler {
	return GetFareHandler{
		distanceCalculator: distanceCalculator,
		directionsService:  directionsService,
	}
}

func (h GetFareHandler) Handle(ctx context.Context, query GetFare) (*domain.Fare, error) {

	directions, err := h.directionsService.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)
	if err != nil || directions == nil {
		directions, err = h.distanceCalculator.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)

		if err != nil {
			return nil, err
		}
	}

	return &domain.Fare{
		ID:                  uuid.Generate(),
		EstimatedDistanceKm: directions.Distance,
		EstimatedDuration:   directions.Duration,
		ETA:                 directions.ArrivalTime,
		Fare:                0.5*directions.Duration.Minutes() + 1.0*directions.Distance/1000,
		Currency:            "EUR",
	}, nil
}
