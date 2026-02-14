package enums

type UserRole string

func (u UserRole) IsValid() bool {
	switch u {
	case UserRoleDriver, UserRoleRider, UserRoleUser, UserRoleAdmin, UserRoleAnonymous:
		return true
	default:
		return false
	}
}

const (
	UserRoleDriver    UserRole = "driver"
	UserRoleRider     UserRole = "rider"
	UserRoleUser      UserRole = "user"
	UserRoleAdmin     UserRole = "admin"
	UserRoleAnonymous UserRole = "anonymous"
)
