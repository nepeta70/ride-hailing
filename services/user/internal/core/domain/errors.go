package domain

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

var (
	// ErrUserNotFound indicates the requested user does not exist.
	// (Pre-existing; kept here for a single source of truth alongside the
	// new auth/profile errors below.)
	ErrUserNotFound = errors.NewErrNotFound("user not found")

	// ErrEmailAlreadyExists indicates a unique constraint violation on email.
	ErrEmailAlreadyExists = errors.NewBusinessError("email already registered")

	// ErrPhoneAlreadyExists indicates a unique constraint violation on phone.
	ErrPhoneAlreadyExists = errors.NewBusinessError("phone already registered")

	// ErrInvalidCredentials indicates a login attempt with a wrong email/password
	// combination. Deliberately generic so callers can't distinguish "wrong
	// password" from "no such email" via error inspection.
	ErrInvalidCredentials = errors.NewBusinessError("invalid email or password")

	// ErrAccountNotActive indicates the account exists but is suspended or deleted.
	ErrAccountNotActive = errors.NewBusinessError("account is not active")

	// ErrProfileNotFound indicates no profile row exists for the given user ID.
	ErrProfileNotFound = errors.NewErrNotFound("user profile not found")
)
