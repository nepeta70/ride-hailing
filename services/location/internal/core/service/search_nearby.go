package service

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/location/internal/core/domain"
)

type SearchNearbyRequest struct {
	Coordinates domain.Coordinates
}

func (r SearchNearbyRequest) Validate() error {
	if err := r.Coordinates.Validate(); err != nil {
		return errors.NewValidationErrorf("invalid coordinates: %w", err)
	}
	return nil
}
