package domain

type UserPayload interface {
	GetUserName() string
	GetUserType() string
	GetFirstName() string
	GetLastName() string
	GetEmail() string
	GetPhoneNumber() string
	GetPassword() string
}

type UserID interface {
	GetUserId() string
}

type UpdateUserPayload interface {
	UserID
	UserPayload
}
