package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type CreateFareRateCommand struct {
	FareRateData
}

func (c *CreateFareRateCommand) Validate() error {
	return c.FareRateData.Validate()
}

type CreateFareRateHandler struct {
	repo ports.FareRatesWriteRepository
}

func NewCreateFareRateHandler(repo ports.FareRatesWriteRepository) *CreateFareRateHandler {
	return &CreateFareRateHandler{repo: repo}
}

func (h *CreateFareRateHandler) Handle(ctx context.Context, cmd *CreateFareRateCommand) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}

	err := cmd.Validate()
	if err != nil {
		return err
	}

	return nil
}
