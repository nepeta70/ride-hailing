package enums

type ServiceName string

const (
	ServiceNameGateway         ServiceName = "gateway"
	ServiceNameUserService     ServiceName = "user"
	ServiceNameDriverService   ServiceName = "driver"
	ServiceNameRiderService    ServiceName = "rider"
	ServiceNameLocationService ServiceName = "location"
	ServiceNameMatching        ServiceName = "matching"
)

func (s ServiceName) IsValid() bool {
	switch s {
	case ServiceNameGateway, ServiceNameUserService, ServiceNameDriverService, ServiceNameRiderService, ServiceNameLocationService, ServiceNameMatching:
		return true
	default:
		return false
	}
}

func (s ServiceName) String() string {
	return string(s)
}
