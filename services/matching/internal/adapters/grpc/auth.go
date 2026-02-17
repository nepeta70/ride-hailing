package grpc

import (
	matchingv1 "github.com/nepeta70/ride-hailing/gen/proto/matching/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type MatchingEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.UserRole
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.UserRole{
		matchingv1.MatchingService_FindMatchingDrivers_FullMethodName: {enums.UserRoleRider, enums.UserRoleDriver, enums.UserRoleAdmin},// TODO: fix roles
	}

	return &MatchingEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *MatchingEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *MatchingEndpointRoles) RequestRoles() map[string][]enums.UserRole {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*MatchingEndpointRoles)(nil)
