package ports

import "github.com/nepeta70/ride-hailing/internal/pkg/domain/enums"

type EndpointRoles interface {
	APIKey() string
	RequestRoles() map[string][]enums.UserRole
}
