package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/user/internal/config"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app/queries"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/validator"
	"github.com/nepeta70/ride-hailing/services/user/internal/ports"
)

type Application struct {
	Commands  *Commands
	Queries   *Queries
	telemetry pkgPorts.TelemetryProvider
	config    *config.Config
}

type Commands struct {
	CreateUser *commands.RegisterUserHandler
	UpdateUser *commands.UpdateUserHandler
}

type Queries struct {
	GetUserByID *queries.GetUserByIDHandler
}

type AppOpts struct {
	Config      *config.Config
	ReadRepo    ports.ReadUserRepository
	WriteRepo   ports.WriteUserRepository
	Credentials ports.UserCredentialsRepository
	Hasher      ports.PasswordHasher
	Telemetry   pkgPorts.TelemetryProvider
	Validator   *validator.PasswordValidator
}

func (o *AppOpts) Validate() error {
	required := []struct {
		name  string
		value any
	}{
		{"config", o.Config},
		{"read repository", o.ReadRepo},
		{"write repository", o.WriteRepo},
		{"credentials repository", o.Credentials},
		{"password hasher", o.Hasher},
		{"telemetry provider", o.Telemetry},
		{"password validator", o.Validator},
	}

	for _, r := range required {
		if r.value == nil {
			return errors.NewValidationErrorf("%s is required", r.name)
		}
	}
	return nil
}

func NewApplication(opts *AppOpts) (*Application, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return &Application{
		Commands: &Commands{
			CreateUser: commands.NewRegisterUserHandler(opts.Credentials, opts.Hasher, opts.Telemetry, opts.Validator),
			UpdateUser: commands.NewUpdateUserHandler(opts.WriteRepo),
		},
		Queries: &Queries{
			GetUserByID: queries.NewGetUserByIDHandler(opts.ReadRepo),
		},
		config:    opts.Config,
		telemetry: opts.Telemetry,
	}, nil
}
