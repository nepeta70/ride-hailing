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
