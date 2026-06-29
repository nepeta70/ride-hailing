package validator

import (
	"regexp"
	"unicode"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

var (
	minLength    = 8
	specialChars = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)
)

type PasswordValidator struct {
	minLength      int
	requireUpper   bool
	requireLower   bool
	requireDigit   bool
	requireSpecial bool
}

func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		minLength:      minLength,
		requireUpper:   true,
		requireLower:   true,
		requireDigit:   true,
		requireSpecial: true,
	}
}

func (v *PasswordValidator) Validate(password string) error {
	if len(password) < v.minLength {
		return errors.NewBusinessErrorf("password must be at least %d characters", v.minLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case specialChars.MatchString(string(char)):
			hasSpecial = true
		}
	}

	if v.requireUpper && !hasUpper {
		return errors.NewBusinessError("password must contain at least one uppercase letter")
	}
	if v.requireLower && !hasLower {
		return errors.NewBusinessError("password must contain at least one lowercase letter")
	}
	if v.requireDigit && !hasDigit {
		return errors.NewBusinessError("password must contain at least one digit")
	}
	if v.requireSpecial && !hasSpecial {
		return errors.NewBusinessError("password must contain at least one special character")
	}

	return nil
}
