package commands

import (
	"context"
	"strings"

	"github.com/google/uuid"
	userv1 "github.com/nepeta70/ride-hailing/gen/proto/user/v1"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/validator"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
	"go.opentelemetry.io/otel/trace"
)

type RegisterUserHandler struct {
	repo              ports.UserCredentialsRepository
	hasher            ports.PasswordHasher
	telemetry         pkgPorts.TelemetryProvider
	passwordValidator validator.PasswordValidator
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd *userv1.RegisterUserRequest) (*domain.UserCredentials, error) {
	ctx, span := h.telemetry.Tracer().Start(ctx, "RegisterUserHandler.Handle",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()
	if err := ctx.Err(); err != nil {
		span.RecordError(err)
		return nil, errors.ErrContextError
	}
	if cmd == nil {
		return nil, errors.NewBusinessError("payload is nil")
	}

	userType, valid := domain.ParseUserType(cmd.GetUserType())
	if !valid {
		return nil, errors.NewBusinessError("invalid user type")
	}

	email := strings.ToLower(strings.TrimSpace(cmd.GetEmail()))
	if email == "" {
		return nil, errors.NewBusinessError("email is required")
	}

	if !validator.IsValidEmail(email) {
		e := errors.NewBusinessErrorf("invalid email format: %s", email)
		span.RecordError(e)
		return nil, e
	}

	phone := strings.TrimSpace(cmd.GetPhoneNumber())
	password := strings.TrimSpace(cmd.GetPassword())
	err := h.passwordValidator.Validate(password)
	if err != nil {
		span.RecordError(err)
		return nil, errors.BusinessError(err)
	}

	passwordHash, err := h.hasher.Hash(cmd.Password)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	credentials := &domain.UserCredentials{
		ID:           uuid.New(),
		Role:         userType,
		Email:        &email,
		Phone:        &phone,
		PasswordHash: passwordHash,
	}

	err = h.repo.Create(ctx, credentials)
	if err != nil {
		span.RecordError(err)
		h.telemetry.Logger().ErrorContext(ctx, "failed to create profile after registering credentials",
			"user_id", credentials.ID, "error", err)

		return nil, err
	}
	return credentials, nil
}

func NewRegisterUserHandler(repo ports.UserCredentialsRepository, hasher ports.PasswordHasher, telemetry pkgPorts.TelemetryProvider, validator *validator.PasswordValidator) *RegisterUserHandler {
	return &RegisterUserHandler{repo: repo, hasher: hasher, telemetry: telemetry, passwordValidator: *validator}
}
