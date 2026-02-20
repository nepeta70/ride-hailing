package grpc

import (
	ridev1 "github.com/nepeta70/ride-hailing/gen/proto/ride/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RideEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.SenderType
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.SenderType{
		ridev1.RideService_EstimateFare_FullMethodName:       {enums.SenderTypeRider},
		ridev1.RideService_RequestRide_FullMethodName:        {enums.SenderTypeRider},
		ridev1.RideService_CancelRide_FullMethodName:         {enums.SenderTypeRider, enums.SenderTypeAdmin},
		ridev1.RideService_AcceptOrRejectRide_FullMethodName: {enums.SenderTypeDriver},
		ridev1.RideService_StartRide_FullMethodName:          {enums.SenderTypeDriver},
		ridev1.RideService_CompleteRide_FullMethodName:       {enums.SenderTypeDriver},
	}

	return &RideEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *RideEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *RideEndpointRoles) RequestRoles() map[string][]enums.SenderType {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*RideEndpointRoles)(nil)
