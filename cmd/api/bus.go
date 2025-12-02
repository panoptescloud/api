package main

import (
	"github.com/panoptescloud/api/internal/common/bus"
	organisationsapp "github.com/panoptescloud/api/internal/organisations/application"
	organisation_users "github.com/panoptescloud/api/internal/organisations/infra/users"
	usersapp "github.com/panoptescloud/api/internal/users/application"
	"github.com/spf13/cobra"
)

var globalBus *bus.Bus

func buildCommandBus() *bus.Bus {
	if globalBus != nil {
		return globalBus
	}

	globalBus = bus.New()

	cobra.CheckErr(usersapp.RegisterToBus(globalBus, svcContainer.GetUsersRepo()))
	cobra.CheckErr(
		organisationsapp.RegisterToBus(
			globalBus,
			svcContainer.GetOrganisationsRepo(),
			svcContainer.GetOrganisationAPIKeysRepo(),
			organisation_users.NewUsersBridge(globalBus),
		),
	)

	return globalBus
}
