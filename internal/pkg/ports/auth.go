package ports

import "github.com/nepeta70/ride-hailing/internal/pkg/core/enums"

type EndpointRoles interface {
	APIKey() string
	RequestRoles() map[string][]enums.SenderType
}
