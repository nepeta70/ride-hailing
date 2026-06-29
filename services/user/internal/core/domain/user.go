package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/validator"
)

type User struct {
	id        uuid.UUID
	userType  Role
	userName  string
	firstName string
	lastName  string
	email     string
	phone     string
	password  string
	createdAt time.Time
	updatedAt time.Time
}

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) UserType() Role       { return u.userType }
func (u *User) UserName() string     { return u.userName }
func (u *User) FirstName() string    { return u.firstName }
func (u *User) LastName() string     { return u.lastName }
func (u *User) Email() string        { return u.email }
func (u *User) Phone() string        { return u.phone }
func (u *User) Password() string     { return u.password }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

func CreateNewUser(payload UserPayload) (*User, error) {
	// 1. Check for nil
	if payload == nil {
		return nil, errors.NewBusinessError("payload is nil")
	}

	userType, valid := ParseUserType(payload.GetUserType())
	if !valid {
		return nil, errors.NewBusinessError("invalid user type")
	}
	// Sanitize input fields
	userName := strings.TrimSpace(payload.GetUserName())
	if userName == "" {
		return nil, errors.NewBusinessError("user name is required")
	}
	firstName := strings.TrimSpace(payload.GetFirstName())
	if firstName == "" {
		return nil, errors.NewBusinessError("first name is required")
	}
	lastName := strings.TrimSpace(payload.GetLastName())
	if lastName == "" {
		return nil, errors.NewBusinessError("last name is required")
	}
	email := strings.ToLower(strings.TrimSpace(payload.GetEmail()))
	if email == "" {
		return nil, errors.NewBusinessError("email is required")
	}
	if !validator.IsValidEmail(email) {
		return nil, errors.NewBusinessError("invalid email format")
	}
	phone := strings.TrimSpace(payload.GetPhoneNumber())
	password := strings.TrimSpace(payload.GetPassword())
	if len(password) < 8 {
		return nil, errors.NewBusinessError("password must be at least 8 characters long")
	}

	user := &User{
		id:        uuid.New(),
		userType:  userType,
		userName:  userName,
		firstName: firstName,
		lastName:  lastName,
		email:     email,
		phone:     phone,
		password:  password,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
	}
	return user, nil
}

func NewUpdateUser(payload UpdateUserPayload) (*User, error) {
	// 1. Check for nil
	if payload == nil {
		return nil, errors.NewBusinessError("payload is nil")
	}
	userId, err := uuid.Parse(payload.GetUserId())
	if err != nil {
		return nil, errors.NewBusinessError("invalid user ID format")
	}
	user, err := CreateNewUser(payload)
	if err != nil {
		return nil, err
	}
	user.id = userId
	return user, nil
}
