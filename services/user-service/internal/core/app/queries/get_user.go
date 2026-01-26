package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/ports"
)

type GetUserByID struct {
	UserID string
}

type GetUserByIDHandler struct {
	readRepo ports.ReadUserRepository
}

func (h GetUserByIDHandler) Handle(ctx context.Context, query GetUserByID) (*domain.User, error) {
	id, err := uuid.Parse(query.UserID)
	if err != nil {
		return nil, errors.BusinessError(err)
	}
	return h.readRepo.GetByID(ctx, id)
}
