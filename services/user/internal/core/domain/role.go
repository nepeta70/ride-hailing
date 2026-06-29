package domain

type Role string

const (
	RoleRider  Role = "rider"
	RoleDriver Role = "driver"
	RoleAdmin  Role = "admin"
)

// String returns the string representation of the user type
func (u Role) String() string {
	return string(u)
}

// Valid checks if the user type is valid
func (u Role) IsValid() bool {
	switch u {
	case RoleRider, RoleDriver, RoleAdmin:
		return true
	default:
		return false
	}
}

// Equals checks if the UserType equals the given string
func (u Role) Equals(str string) bool {
	return string(u) == str
}

// ParseUserType attempts to parse a string into a valid UserType.
// Returns the UserType and true if valid, otherwise returns "" and false.
func ParseUserType(str string) (Role, bool) {
	switch Role(str) {
	case RoleRider, RoleDriver, RoleAdmin:
		return Role(str), true
	default:
		return "", false
	}
}
