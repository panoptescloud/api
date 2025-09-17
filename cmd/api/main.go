package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ErrInvalidOptions = errors.New("invalid options provided")

func handleGroupedCommand(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

var rootCmd = &cobra.Command{
	Use:          "api",
	Short:        "Panoptes API.",
	SilenceUsage: true,
	RunE: handleGroupedCommand,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE:  handleServe,
}

func init() {
	serveCmd.Flags().Int("port", 8080, "The port to serve the API on")

	rootCmd.AddCommand(serveCmd)
}
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
