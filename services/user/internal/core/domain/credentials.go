package domain

import (
	"time"

	"github.com/google/uuid"
)

type CredentialStatus string

const (
	CredentialStatusActive    CredentialStatus = "active"
	CredentialStatusSuspended CredentialStatus = "suspended"
	CredentialStatusDeleted   CredentialStatus = "deleted"
)

func (s CredentialStatus) IsValid() bool {
	switch s {
	case CredentialStatusActive, CredentialStatusSuspended, CredentialStatusDeleted:
		return true
	default:
		return false
	}
}

func (s CredentialStatus) String() string {
	return string(s)
}

// UserCredentials is the auth-critical aggregate backed by user_credentials.
// Kept separate from UserProfile so that access to authentication secrets
// can be controlled and audited independently of profile data.
type UserCredentials struct {
	ID           uuid.UUID
	Email        *string
	Phone        *string
	PasswordHash string
	Role         Role
	Status       CredentialStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CanAuthenticate reports whether the account is in a state that allows
// a login attempt to proceed at all (independent of password correctness).
func (c *UserCredentials) CanAuthenticate(now time.Time) bool {
	return c.Status == CredentialStatusActive
}
