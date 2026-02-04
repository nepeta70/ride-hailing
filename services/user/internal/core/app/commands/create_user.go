package commands

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
)

// 1. The Command (The Data)
type CreateUser struct {
	UserData domain.UserPayload
}

func (c *CreateUser) Validate() error {
	// Add validation logic here

	return nil
}

// 2. The Handler (The Logic)
type CreateUserHandler struct {
	repo ports.WriteUserRepository
}

func (h CreateUserHandler) Handle(ctx context.Context, cmd CreateUser) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.ErrContextError
	}
	user, err := domain.CreateNewUser(cmd.UserData)
	if err != nil {
		return nil, err
	}

	// 2. Persist
	err = h.repo.Save(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func NewCreateUserHandler(repo ports.WriteUserRepository) CreateUserHandler {
	return CreateUserHandler{repo: repo}
}
