package queries

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type GetFareRates struct {
	CountryCode string
	Region      string
}

func (q *GetFareRates) Validate() error {
	if q.CountryCode == "" || q.Region == "" {
		return errors.NewValidationErrorf("country code and region are required")
	}
	if len(q.CountryCode) != 2 {
		return errors.NewValidationErrorf("country code must be 2 characters")
	}

	if len(q.Region) != 0 && len(q.Region) != 2 {
		return errors.NewValidationErrorf("region must be 2 characters")
	}

	return nil
}

type GetFareRateHandler struct {
	repo ports.FareRatesReadRepository
}

func NewGetFareRatesHandler(repo ports.FareRatesReadRepository) *GetFareRateHandler {
	return &GetFareRateHandler{
		repo: repo,
	}
}

func (h *GetFareRateHandler) Handle(ctx context.Context, query *GetFareRates) ([]*domain.FareRate, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}
	err := query.Validate()
	if err != nil {
		return nil, err
	}
	return h.repo.GetRatesByRegion(ctx, query.CountryCode, query.Region)
}
