package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panoptescloud/api/internal/api/auth"
	"github.com/panoptescloud/api/internal/infra/config"
	"github.com/panoptescloud/api/internal/infra/github_oauth"
	"github.com/panoptescloud/api/internal/infra/hasher"
	"github.com/panoptescloud/api/internal/infra/organisations"
	"github.com/panoptescloud/api/internal/infra/repository/postgres"
	authUsers "github.com/panoptescloud/api/internal/infra/users"
	organisationspostgres "github.com/panoptescloud/api/internal/organisations/infra/postgres"
	userspostgres "github.com/panoptescloud/api/internal/users/infra/postgres"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appCfg *config.Config
var cfgFilePath string
var logger *slog.Logger
var svcContainer *services = &services{}

var ErrInvalidOptions = errors.New("invalid options provided")

type services struct {
	githubOauthClient       *github_oauth.Client
	postgresPool            *pgxpool.Pool
	usersRepo               *userspostgres.UsersRepository
	organisationsRepo       *organisationspostgres.OrganisationsRepository
	organisationAPIKeysRepo *organisationspostgres.APIKeysRepository
	sessionManager          *auth.SessionManager
	actorLoader             *auth.ActorLoader
	authOrganisationsBridge *organisations.OrganisationsBridge
	authUsersBridge         *authUsers.UsersBridge
	refreshTokensRepo       *postgres.RefreshTokensRepository
	hasher                  *hasher.HMACSHA256Hasher
}

func (s *services) GetGituhbOauthClient() *github_oauth.Client {
	if s.githubOauthClient != nil {
		return s.githubOauthClient
	}

	s.githubOauthClient = github_oauth.NewClient(
		appCfg.GetGithubOauthClientId(),
		appCfg.GetGithubOauthClientSecret(),
		logger.With("component", "github-oauth-client"),
	)

	return s.githubOauthClient
}

func (s *services) GetPostgresPool() *pgxpool.Pool {
	if s.postgresPool != nil {
		return s.postgresPool
	}

	pool, err := postgres.NewPool(&appCfg.DB.Postgres)

	cobra.CheckErr(err)

	s.postgresPool = pool

	return s.postgresPool
}

func (s *services) GetUsersRepo() *userspostgres.UsersRepository {
	if s.usersRepo != nil {
		return s.usersRepo
	}

	s.usersRepo = userspostgres.NewUsersRepository(
		s.GetPostgresPool(),
	)

	return s.usersRepo
}

func (s *services) GetOrganisationsRepo() *organisationspostgres.OrganisationsRepository {
	if s.organisationsRepo != nil {
		return s.organisationsRepo
	}

	s.organisationsRepo = organisationspostgres.NewOrganisationsRepository(
		s.GetPostgresPool(),
	)

	return s.organisationsRepo
}

func (s *services) GetOrganisationAPIKeysRepo() *organisationspostgres.APIKeysRepository {
	if s.organisationAPIKeysRepo != nil {
		return s.organisationAPIKeysRepo
	}

	s.organisationAPIKeysRepo = organisationspostgres.NewAPIKeysRepository(
		s.GetPostgresPool(),
	)

	return s.organisationAPIKeysRepo
}

func (s *services) GetRefreshTokensRepo() *postgres.RefreshTokensRepository {
	if s.refreshTokensRepo != nil {
		return s.refreshTokensRepo
	}

	s.refreshTokensRepo = postgres.NewRefreshTokensRepository(
		s.GetPostgresPool(),
	)

	return s.refreshTokensRepo
}

func (s *services) GetAuthHasher() *hasher.HMACSHA256Hasher {
	if s.hasher != nil {
		return s.hasher
	}

	s.hasher = hasher.NewHMACSHA256Hasher(
		[]byte(appCfg.GetAuthHMACKey()),
	)

	return s.hasher
}

func (s *services) GetSessionManager() *auth.SessionManager {
	if s.sessionManager != nil {
		return s.sessionManager
	}

	svc, err := auth.NewSessionManager(
		s.GetRefreshTokensRepo(),
		s.GetAuthHasher(),
		appCfg.GetAuthJWTPrivateKeyPath(),
		appCfg.GetAuthJWTPublicKeyPath(),
	)

	cobra.CheckErr(err)

	s.sessionManager = svc

	return s.sessionManager
}

func (s *services) GetAuthOrganisationsBridge() *organisations.OrganisationsBridge {
	if s.authOrganisationsBridge != nil {
		return s.authOrganisationsBridge
	}

	s.authOrganisationsBridge = organisations.NewOrganisationsBridge(globalBus)

	return s.authOrganisationsBridge
}

func (s *services) GetAuthUsersBridge() *authUsers.UsersBridge {
	if s.authUsersBridge != nil {
		return s.authUsersBridge
	}

	s.authUsersBridge = authUsers.NewUsersBridge(globalBus)

	return s.authUsersBridge
}

func (s *services) GetActorLoader() *auth.ActorLoader {
	if s.actorLoader != nil {
		return s.actorLoader
	}

	s.actorLoader = auth.NewActorLoader(
		s.GetAuthOrganisationsBridge(),
		s.GetAuthUsersBridge(),
	)

	return s.actorLoader
}

func handleGroupedCommand(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

var rootCmd = &cobra.Command{
	Use:          "api",
	Short:        "Panoptes API.",
	SilenceUsage: true,
	RunE:         handleGroupedCommand,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE:  handleServe,
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Commands to aid in debugging",
	RunE:  handleGroupedCommand,
}

var debugShowConfigCmd = &cobra.Command{
	Use:   "show-config",
	Short: "Shows the currently loaded configuration.",
	Long:  `Includes any overrides provided by environent variabels or cli flags.`,
	RunE:  handleDebugShowConfig,
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

// bindVipersEnvsFromStruct is a bit of a workaround for dodgy viper behaviour.
// One would assume that "AutomaticEnv" does this, but thats not actually very
// automatic, and we still need to bind each env to the relevant key aparrently.
// This does so by recursing through the config struct and binding each key.
func bindVipersEnvsFromStruct(prefix string, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// use yaml tag if present, otherwise field name
		key := field.Tag.Get("yaml")
		if key == "" {
			key = strings.ToLower(field.Name)
		}

		if prefix != "" {
			key = prefix + "." + key
		}

		if field.Type.Kind() == reflect.Struct {
			bindVipersEnvsFromStruct(key, field.Type)
		} else {
			_ = viper.BindEnv(key)
		}
	}
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

	bindVipersEnvsFromStruct("", reflect.TypeOf(config.Config{}))

	appCfg = config.Default()

	err = viper.Unmarshal(appCfg)
	cobra.CheckErr(err)
}

func configureLogger() {
	l := slog.Level(slog.LevelInfo)
	err := l.UnmarshalText([]byte(appCfg.Logging.Level))

	cobra.CheckErr(err)

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     l,
	}

	if appCfg.Logging.Format == "text" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	logger = logger.With(slog.String("service", "panoptes"))
	slog.SetDefault(logger)
}

func bootstrap() {
	loadConfig()
	configureLogger()
	buildCommandBus()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
