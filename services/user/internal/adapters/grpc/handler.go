package grpc

import (
	"context"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	application *app.Application
}

func NewUserHandler(application *app.Application) *UserHandler {
	return &UserHandler{
		application: application,
	}
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.User, error) {

	user, err := h.application.Queries.GetUserByID.Handle(ctx, queries.GetUserByID{UserID: req.GetUserId()})
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return &userv1.User{
		UserId:      user.ID().String(),
		UserType:    user.UserType().String(),
		UserName:    user.UserName(),
		FirstName:   user.FirstName(),
		LastName:    user.LastName(),
		Email:       user.Email(),
		PhoneNumber: user.Phone(),
	}, nil
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
	user, err := h.application.Commands.CreateUser.Handle(ctx, req)
	if err != nil {
		return nil, err
	}

	return &userv1.RegisterUserResponse{
		UserId: user.ID.String(),
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.User, error) {
	err := h.application.Commands.UpdateUser.Handle(ctx, req)

	if err != nil {
		return nil, err
	}
	return &userv1.User{
		UserId:    req.GetUserId(),
		UserName:  req.GetUserName(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}, nil
}
