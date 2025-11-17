package main

import (
	"github.com/panoptescloud/api/internal/common/bus"
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

	return globalBus
}
