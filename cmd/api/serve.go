package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	stdhttp "net/http"

	"github.com/panoptescloud/api/internal/api/http"
	"github.com/panoptescloud/api/internal/api/http/v1beta"
	"github.com/panoptescloud/api/internal/infra/github_oauth"
	"github.com/spf13/cobra"
)

func handleServe(_ *cobra.Command, _ []string) error {
	api := http.NewServer(
		svcContainer.GetSessionManager(),
		logger.With("component", "http-server"),
	)

	controllers := []http.Controller{
		v1beta.NewProbesController(),
		v1beta.NewUsersController(),
		v1beta.NewAuthController(
			globalBus,
			github_oauth.NewClient(
				appCfg.GetGithubOauthClientId(),
				appCfg.GetGithubOauthClientSecret(),
				logger.With("component", "github-oauth-client"),
			),
			svcContainer.GetSessionManager(),
			logger.With("component", "auth-controller.v1beta"),
		),
	}

	api.Initialise(controllers)

	serverErrors := make(chan error, 1)

	go func() {
		// Check for stdhttp.ErrServerClosed as this is what is returned when
		// shutdown is called, the start method should return this at that point,
		// but it's not an error we actually want to handle, as it's part of our
		// graceful shutdown.
		if err := api.Start(uint16(appCfg.GetServerPort())); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		// Fallthrough to code after this select, to initiate graceful shutdown
	case err := <-serverErrors:
		//	Something went wrong during startup
		return err
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return api.Shutdown(ctx)
}
