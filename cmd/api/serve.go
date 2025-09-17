package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

func handleServe(_ *cobra.Command, _ []string) error {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	return e.Start(fmt.Sprintf(":%d", appCfg.GetServerPort()))
}