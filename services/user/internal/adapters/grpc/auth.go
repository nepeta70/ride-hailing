package grpc

import (
	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type UserEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.SenderType
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.SenderType{
		userv1.UserService_GetUser_FullMethodName:      {enums.SenderTypeUser, enums.SenderTypeAdmin, enums.SenderTypeService},
		userv1.UserService_RegisterUser_FullMethodName: {enums.SenderTypeAnonymous},
		userv1.UserService_UpdateUser_FullMethodName:   {enums.SenderTypeUser, enums.SenderTypeAdmin},
	}

	return &UserEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *UserEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *UserEndpointRoles) RequestRoles() map[string][]enums.SenderType {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*UserEndpointRoles)(nil)
