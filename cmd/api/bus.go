package main

import (
	"github.com/panoptescloud/api/internal/application/bus"
	"github.com/panoptescloud/api/internal/application/users"
)

var globalBus *bus.Bus

func buildCommandBus() *bus.Bus {
	if globalBus != nil {
		return globalBus
	}

	globalBus = bus.New()

	// --- Users
	userQueryHandler := users.NewUserQueryHandler(
		svcContainer.GetUsersRepo(),
	)

	userCmdHandler := users.NewUserCmdHandler(
		svcContainer.GetUsersRepo(),
	)

	bus.RegisterCommand(globalBus, userCmdHandler.CreateUser)
	bus.RegisterQuery(globalBus, userQueryHandler.GetUserByGithubNodeID)

	return globalBus
}
