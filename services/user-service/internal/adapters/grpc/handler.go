package grpc

import (
	"context"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/service"
)

// UserHandler implements the UserService gRPC interface.
type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.User, error) {
	// TODO: Call the user service logic here
	return &userv1.User{}, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.User, error) {
	// TODO: Call the user service logic here
	return &userv1.User{}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.User, error) {
	// TODO: Call the user service logic here
	return &userv1.User{}, nil
}
