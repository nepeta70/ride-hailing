package domain

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Password  string // hashed password
	CreatedAt int64  // Unix timestamp
	UpdatedAt int64  // Unix timestamp

}
type DriverProfile struct {
	UserID        string
	LicenseNumber string
	VehicleInfo   string
	Rating        float32
	CreatedAt     int64 // Unix timestamp
	UpdatedAt     int64 // Unix timestamp
}
