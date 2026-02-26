package commands

import (
	"context"
	"math"

	"github.com/google/uuid"
	core "github.com/nepeta70/ride-hailing/internal/pkg/core"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type EstimateFareCommand struct {
	RiderID         uuid.UUID
	RequestID       uuid.UUID
	PickupLocation  *core.Coordinates
	DropoffLocation *core.Coordinates
	CountryCode     string
}

func (q *EstimateFareCommand) Validate() error {
	if q.RiderID == uuid.Nil {
		return errors.NewValidationErrorf("invalid rider ID format")
	}

	err := q.PickupLocation.Validate()
	if err != nil {
		return errors.NewValidationErrorf("invalid pickup location: %v", err)
	}
	err = q.DropoffLocation.Validate()
	if err != nil {
		return errors.NewValidationErrorf("invalid dropoff location: %v", err)
	}

	if q.RequestID == uuid.Nil {
		return errors.NewValidationErrorf("invalid request ID format")
	}
	return nil
}

type EstimateFareHandler struct {
	config            *config.Config
	countryCache      ports.CountryCacheInterface
	directionsService ports.DirectionsService
	storage           ports.StorageBundle
	logger            pkgPorts.Logger
}

func NewEstimateFareHandler(config *config.Config, logger pkgPorts.Logger, storage ports.StorageBundle, directionsService ports.DirectionsService) *EstimateFareHandler {
	return &EstimateFareHandler{
		config:            config,
		countryCache:      storage.CountryCache(),
		directionsService: directionsService,
		storage:           storage,
		logger:            logger,
	}
}

func (h *EstimateFareHandler) Handle(ctx context.Context, query EstimateFareCommand) (*domain.Fares, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}

	err := query.Validate()
	if err != nil {
		return nil, err
	}

	// Idempotent check - if fare estimate already exists for this request, return it
	f, err := h.storage.FareReadRepo().GetByID(ctx, query.RequestID)
	if f != nil {
		return f, nil
	}

	country, exists := h.countryCache.GetCountryByCode(ctx, query.CountryCode)
	if !exists {
		h.logger.Error("country not found", "country_code", query.CountryCode)
		return nil, errors.NewErrNotFound("country not found")
	}

	fareRates, err := h.storage.FareRatesReadRepo().GetRatesByCountry(ctx, query.CountryCode)
	if err != nil {
		return nil, err
	}

	n := len(fareRates)
	if n == 0 {
		h.logger.Error("no fare rates found for country", "country_code", query.CountryCode)
		return nil, errors.NewErrNotFound("no fare rates found for country")
	}

	directions, err := h.directionsService.GetDirections(ctx, query.PickupLocation, query.DropoffLocation)
	if err != nil {
		h.logger.Error("failed to get directions", "error", err)
		return nil, errors.NewPermanentErrorf("failed to get directions: %w", err)
	}

	fares := make(map[string]float64, n)
	distanceInKm := directions.DistanceMeters / 1000.0
	for _, rate := range fareRates {
		fareAmount := rate.BaseFare + (rate.FarePerKm * distanceInKm) + (rate.FarePerMinute * directions.DurationMinutes.Minutes())
		fares[rate.ServiceType] = math.Max(rate.MinimumFare, fareAmount)
	}

	record := &domain.Fares{
		ID:                       query.RequestID,
		RequestID:                query.RequestID,
		PickupLocation:           query.PickupLocation,
		DropoffLocation:          query.DropoffLocation,
		EstimatedDistanceKm:      distanceInKm,
		EstimatedDurationMinutes: directions.DurationMinutes,
		ETA:                      directions.ArrivalTime,
		Currency:                 country.Currency,
		Fares:                    fares,
	}

	err = h.storage.FareWriteRepo().Save(ctx, record)
	if err != nil {
		h.logger.Error("failed to save fare estimate", "error", err)
		return nil, errors.NewPermanentErrorf("failed to save fare estimate: %w", err)
	}
	return record, nil
}
