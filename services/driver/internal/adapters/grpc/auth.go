package grpc

import (
	driverv1 "github.com/nepeta70/ride-hailing/gen/proto/driver/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type DriverEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.UserRole
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.UserRole{
		driverv1.DriverService_CreateDriver_FullMethodName: {enums.UserRoleAdmin},
		driverv1.DriverService_UpdateDriver_FullMethodName: {enums.UserRoleDriver, enums.UserRoleAdmin},
		driverv1.DriverService_GetDriver_FullMethodName:    {enums.UserRoleDriver, enums.UserRoleAdmin},
	}

	return &DriverEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *DriverEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *DriverEndpointRoles) RequestRoles() map[string][]enums.UserRole {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*DriverEndpointRoles)(nil)
