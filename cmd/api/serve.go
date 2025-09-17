package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

func handleServe(cmd *cobra.Command, _ []string) error {
	port, err := cmd.Flags().GetInt("port")
	cobra.CheckErr(err)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	return e.Start(fmt.Sprintf(":%d", port))
}