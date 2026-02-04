package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
)

type WriteUserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
}

type ReadUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
