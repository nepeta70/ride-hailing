package domain

type CreateUserRequest struct {
	UserType  string
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Password  string
}
