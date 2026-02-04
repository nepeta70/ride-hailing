package domain

type UserType string

const (
	UserTypeRider  UserType = "rider"
	UserTypeDriver UserType = "driver"
)

// String returns the string representation of the user type
func (u UserType) String() string {
	return string(u)
}

// Valid checks if the user type is valid
func (u UserType) Valid() bool {
	switch u {
	case UserTypeRider, UserTypeDriver:
		return true
	default:
		return false
	}
}

// Equals checks if the UserType equals the given string
func (u UserType) Equals(str string) bool {
	return string(u) == str
}

// ParseUserType attempts to parse a string into a valid UserType.
// Returns the UserType and true if valid, otherwise returns "" and false.
func ParseUserType(str string) (UserType, bool) {
	switch UserType(str) {
	case UserTypeRider, UserTypeDriver:
		return UserType(str), true
	default:
		return "", false
	}
}
