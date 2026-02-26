package grpc

import (
	locationv1 "github.com/nepeta70/ride-hailing/gen/proto/location/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type LocationEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.SenderType
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.SenderType{
		locationv1.LocationService_DeleteDriverLocation_FullMethodName: {enums.SenderTypeDriver},
		locationv1.LocationService_SearchNearbyDrivers_FullMethodName:  {enums.SenderTypeAdmin, enums.SenderTypeService, enums.SenderTypeRider},
		locationv1.LocationService_UpdateDriverLocation_FullMethodName: {enums.SenderTypeDriver},
	}

	return &LocationEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *LocationEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *LocationEndpointRoles) RequestRoles() map[string][]enums.SenderType {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*LocationEndpointRoles)(nil)
