package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/config"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app/queries"
)

type Application struct {
	Commands *Commands
	Queries  *Queries
	logger   ports.Logger
	config   *config.Config
}

type Commands struct {
	CreateUser commands.CreateUserHandler
	UpdateUser commands.UpdateUserHandler
}

type Queries struct {
	GetUserByID queries.GetUserByIDHandler
}

func NewApplication(cfg *config.Config, logger ports.Logger) *Application {
	return &Application{
		Commands: &Commands{},
		Queries:  &Queries{},
		logger:   logger,
		config:   cfg,
	}
}
