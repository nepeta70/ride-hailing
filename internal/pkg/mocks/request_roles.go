package mocks

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type EndpointRequests struct{}

func (m *EndpointRequests) APIKey() string {
	return "test-secret"
}

func (m *EndpointRequests) RequestRoles() map[string][]enums.UserRole {
	return map[string][]enums.UserRole{}
}

var _ ports.EndpointRoles = (*EndpointRequests)(nil)
