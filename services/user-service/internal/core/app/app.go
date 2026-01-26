package app

import (
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app/commands"
	"github.com/nepeta70/ride-hailing/services/user-service/internal/core/app/queries"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateUser commands.CreateUserHandler
	UpdateUser commands.UpdateUserHandler
}

type Queries struct {
	GetUserByID queries.GetUserByIDHandler
}
