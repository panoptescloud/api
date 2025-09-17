package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

func handleDebugShowConfig(cmd *cobra.Command, _ []string) error {
	b, err := yaml.Marshal(appCfg)
	if err != nil {
		return err
	}

	fmt.Print(string(b))

	return nil
}