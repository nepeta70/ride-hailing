package pgstore

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	"github.com/nepeta70/ride-hailing/services/user/internal/config"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
)

type CredentialsRepo struct {
	config *config.Config
	db     *pgstore.PostgresDB
}

func NewCredentialsRepo(config *config.Config, db *pgstore.PostgresDB) *CredentialsRepo {
	return &CredentialsRepo{
		config: config,
		db:     db,
	}
}

func (r *CredentialsRepo) Create(ctx context.Context, creds *domain.UserCredentials) error {
	const query = `
		INSERT INTO user_credentials (
			id, email, phone, role, password_hash
		) VALUES ($1, $2, $3, $4, $5)
	`
	var email any
	if creds.Email != nil {
		email = strings.ToLower(*creds.Email)
	}

	var phone any
	if creds.Phone != nil {
		phone = *creds.Phone
	}

	_, err := r.db.ExecContext(ctx, query,
		creds.ID, email, phone, creds.Role.String(), creds.PasswordHash,
	)
	if err != nil {
		return mapUniqueConstraintError(err)
	}
	return nil
}

func (r *CredentialsRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserCredentials, error) {
	const query = `
		SELECT id, email, phone, role, status, created_at, updated_at
		FROM user_credentials
		WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

func (r *CredentialsRepo) GetByEmail(ctx context.Context, email string) (*domain.UserCredentials, error) {
	const query = `
		SELECT id, email, phone, role, status, created_at, updated_at
		FROM user_credentials
		WHERE email = $1
	`
	return r.scanOne(ctx, query, strings.ToLower(email))
}

func (r *CredentialsRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CredentialStatus) error {
	const query = `
		UPDATE user_credentials
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id, status.String())
	if err != nil {
		return err
	}
	return checkRowsAffected(result, domain.ErrUserNotFound)
}

func (r *CredentialsRepo) scanOne(ctx context.Context, query string, arg any) (*domain.UserCredentials, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, domain.ErrUserNotFound
	}

	var e CredentialsEntity
	if err := rows.Scan(
		&e.ID, &e.Email, &e.Phone, &e.Role, &e.Status, &e.CreatedAt, &e.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return e.ToDomain(), nil
}

func checkRowsAffected(result sql.Result, notFoundErr error) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notFoundErr
	}
	return nil
}

func mapUniqueConstraintError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "user_credentials_email_key"):
		return domain.ErrEmailAlreadyExists
	case strings.Contains(msg, "user_credentials_phone_key"):
		return domain.ErrPhoneAlreadyExists
	default:
		return errors.NewTransientErrorf("database error during credentials write: %w", err)
	}
}

var _ ports.UserCredentialsRepository = (*CredentialsRepo)(nil)
