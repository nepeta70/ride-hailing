package grpc

import (
	notificationv1 "github.com/nepeta70/ride-hailing/gen/proto/notification/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type NotificationEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.UserRole
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.UserRole{
		notificationv1.NotificationService_SendNotification_FullMethodName: {enums.UserRoleAdmin, enums.UserRoleUser}, // TODO: Adjust roles
	}

	return &NotificationEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *NotificationEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *NotificationEndpointRoles) RequestRoles() map[string][]enums.UserRole {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*NotificationEndpointRoles)(nil)
