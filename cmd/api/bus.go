package main

import (
	"github.com/panoptescloud/api/internal/application/bus"
	"github.com/panoptescloud/api/internal/application/users"
	"github.com/panoptescloud/api/internal/infra/github_oauth"
)

var globalBus *bus.Bus

func buildCommandBus() *bus.Bus {
	if globalBus != nil {
		return globalBus
	}

	userHandlers := users.NewUserHandlers(
		github_oauth.NewClient(
			appCfg.GetGithubOauthClientId(),
			appCfg.GetGithubOauthClientSecret(),
			logger.With("component", "github-oauth-client"),
		),
	)

	globalBus = bus.New()

	bus.RegisterCommand(globalBus, userHandlers.SigninOrRegisterViaGithub)

	return globalBus
}
