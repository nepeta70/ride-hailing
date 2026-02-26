package mocks

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type EndpointRequests struct{}

func (m *EndpointRequests) APIKey() string {
	return "test-secret"
}

func (m *EndpointRequests) RequestRoles() map[string][]enums.SenderType {
	return map[string][]enums.SenderType{}
}

var _ ports.EndpointRoles = (*EndpointRequests)(nil)
