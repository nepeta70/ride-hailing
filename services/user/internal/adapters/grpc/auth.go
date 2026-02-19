package grpc

import (
	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type UserEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.UserRole
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.UserRole{
		userv1.UserService_GetUser_FullMethodName:    {enums.UserRoleUser, enums.UserRoleAdmin},
		userv1.UserService_CreateUser_FullMethodName: {enums.UserRoleAdmin},
		userv1.UserService_UpdateUser_FullMethodName: {enums.UserRoleUser, enums.UserRoleAdmin},
	}

	return &UserEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *UserEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *UserEndpointRoles) RequestRoles() map[string][]enums.UserRole {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*UserEndpointRoles)(nil)
