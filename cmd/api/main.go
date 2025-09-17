package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/panoptescloud/api/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appCfg *config.Config
var cfgFilePath string

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

var debugCmd = &cobra.Command{
	Use: "debug",
	Short: "Commands to aid in debugging",
	RunE: handleGroupedCommand,
}

var debugShowConfigCmd = &cobra.Command{
	Use: "show-config",
	Short: "Shows the currently loaded configuration.",
	Long: `Includes any overrides provided by environent variabels or cli flags.`,
	RunE: handleDebugShowConfig,
}

func init() {
	cobra.OnInitialize(bootstrap)

	currentDir, err := os.Getwd()
	cobra.CheckErr(err)
	defaultCfgFilePath := fmt.Sprintf("%s/api.panoptes.yaml", currentDir)

	rootCmd.PersistentFlags().StringVar(&cfgFilePath, "config", defaultCfgFilePath, "Path to config file to use")
	rootCmd.PersistentFlags().String("log-level", "error", "log level to use")
	rootCmd.PersistentFlags().String("log-format", "json", "log format to use")
	
	serveCmd.Flags().Int("port", 8080, "The port to serve the API on")

	rootCmd.AddCommand(serveCmd)

	debugCmd.AddCommand(debugShowConfigCmd)
	rootCmd.AddCommand(debugCmd)

	cobra.CheckErr(viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level")))
	cobra.CheckErr(viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("log-format")))
	cobra.CheckErr(viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port")))
}

func loadConfig() {
	// Tell viper to replace . in nested path with underscores
	// e.g. logging.level becomes LOGGING_LEVEL
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetEnvPrefix("panoptes")
	viper.AutomaticEnv()
	viper.SetConfigFile(cfgFilePath)
	

	_, err := os.Stat(cfgFilePath)

	// TODO: better error handling here
	cobra.CheckErr(err)

	err = viper.ReadInConfig()
	cobra.CheckErr(err)

	appCfg = config.Default()

	err = viper.Unmarshal(appCfg)
	cobra.CheckErr(err)
}

func bootstrap() {
	loadConfig()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
