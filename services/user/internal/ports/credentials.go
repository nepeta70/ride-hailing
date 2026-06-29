package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
)

type UserCredentialsRepository interface {
	Create(ctx context.Context, creds *domain.UserCredentials) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.UserCredentials, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserCredentials, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CredentialStatus) error
}

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Verify(hash, plaintext string) (bool, error)
}
