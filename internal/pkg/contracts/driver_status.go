package contracts

type DriverStatus string

const (
	DriverStatusAvailable DriverStatus = "available"
	DriverStatusBusy      DriverStatus = "busy"
	DriverStatusOffline   DriverStatus = "offline"
)

func (s DriverStatus) IsValid() bool {
	switch s {
	case DriverStatusAvailable, DriverStatusBusy, DriverStatusOffline:
		return true
	default:
		return false
	}
}

func (s DriverStatus) Equals(other string) bool {
	return string(s) == other
}

func (s DriverStatus) String() string {
	return string(s)
}
