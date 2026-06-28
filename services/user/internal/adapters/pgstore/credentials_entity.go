package pgstore

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
)

// CredentialsEntity is the database row mapping for user_credentials.
// Private to the adapter and only used for SQL scanning.
type CredentialsEntity struct {
	ID           uuid.UUID      `db:"id"`
	Email        sql.NullString `db:"email"`
	Phone        sql.NullString `db:"phone"`
	PasswordHash string         `db:"password_hash"`
	Role         string         `db:"role"`
	Status       string         `db:"status"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

func (e *CredentialsEntity) ToDomain() *domain.UserCredentials {
	var emailPtr *string
	if e.Email.Valid {
		emailVal := e.Email.String
		emailPtr = &emailVal
	}

	var phonePtr *string
	if e.Phone.Valid {
		phoneVal := e.Phone.String
		phonePtr = &phoneVal
	}

	return &domain.UserCredentials{
		ID:        e.ID,
		Email:     emailPtr,
		Phone:     phonePtr,
		Role:      domain.Role(e.Role),
		Status:    domain.CredentialStatus(e.Status),
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
