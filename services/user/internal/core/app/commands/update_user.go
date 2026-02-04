package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
)

// 1. The Command (The Data)
type UpdateUser struct {
	UserData domain.UpdateUserPayload
}

// 2. The Handler (The Logic)
type UpdateUserHandler struct {
	repo ports.WriteUserRepository
}

func (h UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUser) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}
	user, err := domain.NewUpdateUser(cmd.UserData)
	if err != nil {
		return err
	}

	// 2. Persist
	return h.repo.Update(ctx, user)
}
func NewUpdateUserHandler(repo ports.WriteUserRepository) UpdateUserHandler {
	return UpdateUserHandler{repo: repo}
}
