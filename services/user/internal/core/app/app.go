package app

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/user/internal/config"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/user/internal/core/app/queries"
	userPorts "github.com/nepeta70/ride-hailing/services/user/internal/ports"
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

func newCommands(writeRepo userPorts.WriteUserRepository) *Commands {
	return &Commands{
		CreateUser: commands.NewCreateUserHandler(writeRepo),
		UpdateUser: commands.NewUpdateUserHandler(writeRepo),
	}
}

type Queries struct {
	GetUserByID queries.GetUserByIDHandler
}

func newQueries(readRepo userPorts.ReadUserRepository) *Queries {
	return &Queries{
		GetUserByID: queries.NewGetUserByIDHandler(readRepo),
	}
}

func NewApplication(cfg *config.Config, logger ports.Logger, readRepo userPorts.ReadUserRepository, writeRepo userPorts.WriteUserRepository) *Application {
	app := &Application{
		Commands: newCommands(writeRepo),
		Queries:  newQueries(readRepo),
		logger:   logger,
		config:   cfg,
	}
	return app
}
