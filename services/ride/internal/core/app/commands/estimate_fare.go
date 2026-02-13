package commands

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type EstimateFareCommand struct {
	RiderID         string
	RequestID       string
	PickupLocation  string
	DropoffLocation string
	CountryCode     string
}

func (q *EstimateFareCommand) Validate() error {
	err := uuid.Validate(q.RiderID)
	if err != nil {
		return errors.NewValidationErrorf("invalid rider ID format")
	}

	if q.PickupLocation == "" || q.DropoffLocation == "" {
		return errors.NewValidationErrorf("pickup and dropoff locations are required")
	}
	if q.PickupLocation == q.DropoffLocation {
		return errors.NewValidationErrorf("pickup and dropoff locations cannot be the same")
	}
	err = uuid.Validate(q.RequestID)
	if err != nil {
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

	requestID := uuid.MustParse(query.RequestID)
	// Idempotent check - if fare estimate already exists for this request, return it
	f, err := h.storage.FareReadRepo().GetByID(ctx, requestID)
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
		ID:                       requestID,
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
