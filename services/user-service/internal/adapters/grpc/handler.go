package grpc

import (
	"context"

	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/domain"
)

// UserHandler implements the UserService gRPC interface.
type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	createUserCommandHandler commands.CreateUserHandler
	updateUserCommandHandler commands.UpdateUserHandler
	getUserQueryHandler      queries.GetUserByIDHandler
}

func NewUserHandler(createUserCommandHandler commands.CreateUserHandler, updateUserCommandHandler commands.UpdateUserHandler, getUserQueryHandler queries.GetUserByIDHandler) *UserHandler {
	return &UserHandler{
		createUserCommandHandler: createUserCommandHandler,
		updateUserCommandHandler: updateUserCommandHandler,
		getUserQueryHandler:      getUserQueryHandler,
	}
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.User, error) {

	user, err := h.getUserQueryHandler.Handle(ctx, queries.GetUserByID{UserID: req.GetUserId()})
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return &userv1.User{
		UserId:      user.ID().String(),
		FirstName:   user.FirstName(),
		LastName:    user.LastName(),
		Email:       user.Email(),
		PhoneNumber: user.Phone(),
	}, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.User, error) {

	var command = commands.CreateUser{UserData: req}

	user, err := h.createUserCommandHandler.Handle(ctx, command)
	if err != nil {
		return nil, err
	}

	return &userv1.User{
		UserId:      user.ID().String(),
		UserType:    user.UserType().String(),
		FirstName:   user.FirstName(),
		LastName:    user.LastName(),
		Email:       user.Email(),
		PhoneNumber: user.Phone(),
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.User, error) {
	var command = commands.UpdateUser{UserData: req}

	err := h.updateUserCommandHandler.Handle(ctx, command)

	if err != nil {
		return nil, err
	}
	return &userv1.User{
		UserId:      req.GetUserId(),
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		Email:       req.GetEmail(),
		PhoneNumber: req.GetPhoneNumber(),
	}, nil
}
