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

	userHandlers := users.NewUserHandlers(
		svcContainer.GetGituhbOauthClient(),
		svcContainer.GetUsersRepo(),
	)

	globalBus = bus.New()

	bus.RegisterQuery(globalBus, userHandlers.GetUserByGithubNodeID)

	return globalBus
}
