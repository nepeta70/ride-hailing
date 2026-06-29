package commands

import (
	"context"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
)

type UpdateUserHandler struct {
	repo ports.WriteUserRepository
}

func (h UpdateUserHandler) Handle(ctx context.Context, cmd *userv1.UpdateUserRequest) error {
	if err := ctx.Err(); err != nil {
		return errors.ErrContextError
	}
	// user, err := domain.NewUpdateUser(cmd.UserData)
	// if err != nil {
	// 	return err
	// }

	// // 2. Persist
	// return h.repo.Update(ctx, user)
	return nil // TODO implement
}
func NewUpdateUserHandler(repo ports.WriteUserRepository) *UpdateUserHandler {
	return &UpdateUserHandler{repo: repo}
}
