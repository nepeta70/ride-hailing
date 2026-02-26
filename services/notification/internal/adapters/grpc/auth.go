package grpc

import (
	notificationv1 "github.com/nepeta70/ride-hailing/gen/proto/notification/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/config"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type NotificationEndpointRoles struct {
	apiKey       string
	requestRoles map[string][]enums.SenderType
}

func NewEndpointRoles(config *config.BaseConfig) ports.EndpointRoles {
	var roleConfig = map[string][]enums.SenderType{
		notificationv1.NotificationService_SendNotification_FullMethodName: {enums.SenderTypeService},
	}

	return &NotificationEndpointRoles{
		apiKey:       config.APIKey,
		requestRoles: roleConfig,
	}
}

func (c *NotificationEndpointRoles) APIKey() string {
	return c.apiKey
}

func (c *NotificationEndpointRoles) RequestRoles() map[string][]enums.SenderType {
	return c.requestRoles
}

var _ ports.EndpointRoles = (*NotificationEndpointRoles)(nil)
