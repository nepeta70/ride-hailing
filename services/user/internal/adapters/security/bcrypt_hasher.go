package security

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = bcrypt.DefaultCost // 10; bump if hardware allows

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: defaultBcryptCost}
}

func (h *BcryptHasher) Hash(plaintext string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), h.cost)
	if err != nil {
		return "", errors.NewPermanentErrorf("failed to hash password: %w", err)
	}
	return string(hashed), nil
}

func (h *BcryptHasher) Verify(hash, plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, errors.NewPermanentErrorf("failed to verify password: %w", err)
}

var _ ports.PasswordHasher = (*BcryptHasher)(nil)
