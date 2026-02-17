package grpc

import (
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type LocationEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.UserRole
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.UserRole{
		locationv1.LocationService_GetUserLocation_FullMethodName:     {enums.UserRoleRider},
		locationv1.LocationService_SearchNearbyDrivers_FullMethodName: {enums.UserRoleDriver},
		locationv1.LocationService_UpdateUserLocation_FullMethodName:  {enums.UserRoleRider, enums.UserRoleDriver},
	}

	return &LocationEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *LocationEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *LocationEndpointRoles) RequestRoles() map[string][]enums.UserRole {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*LocationEndpointRoles)(nil)
