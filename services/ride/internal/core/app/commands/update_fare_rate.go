package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type UpdateFareRateCommand struct {
	FareRateData
}

func (c *UpdateFareRateCommand) Validate() error {
	return c.FareRateData.Validate()
}

type UpdateFareRateHandler struct {
	repo ports.FareRatesWriteRepository
}

func NewUpdateFareRateHandler(repo ports.FareRatesWriteRepository) *UpdateFareRateHandler {
	return &UpdateFareRateHandler{repo: repo}
}

func (h *UpdateFareRateHandler) Handle(ctx context.Context, cmd *UpdateFareRateCommand) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	err := cmd.Validate()
	if err != nil {
		return err
	}

	return nil
}
