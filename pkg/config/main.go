package config

type PostgresSSLMode string
const (
	PostgresSSLModeDisable PostgresSSLMode = "disable"
	PostgresSSLModeAllow PostgresSSLMode = "allow"
	PostgresSSLModePrefer PostgresSSLMode = "prefer"
	PostgresSSLModeRequire PostgresSSLMode = "require"
	PostgresSSLModeVerifyCA PostgresSSLMode = "verify-ca"
	PostgresSSLModeVerifyFull PostgresSSLMode = "verify-full"
)

type ServerAccessLogsConfig struct {
	Format  string
	Enabled bool
}

type ServerConfig struct {
	Port       uint16
	AccessLogs ServerAccessLogsConfig `yaml:"accessLogs"`
}

type GithubOauthConfig struct {
	ClientId     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

type GithubConfig struct {
	Oauth GithubOauthConfig `yaml:"oauth"`
}

type LoggingConfig struct {
	Level  string
	Format string
}

type PostgresConfig struct {
	Host string
	Username string
	Password string
	Port uint16
	DBName string `yaml:"dbName"`
	MaxConnections int32 `yaml:"maxConnections"`
	MinConnections int32 `yaml:"minConnections"`
	SSLMode PostgresSSLMode `yaml:"sslMode"`
}

type DBConfig struct {
	Postgres PostgresConfig `yaml:"postgres"`
}

type JWTConfig struct {
	PrivateKeyFile string `yaml:"privateKeyFile"`
	PublicKeyFile string `yaml:"publicKeyFile"`
}

type AuthConfig struct {
	JWT JWTConfig `yaml:"jwt"`
}

type Config struct {
	Server  ServerConfig
	Logging LoggingConfig
	Github  GithubConfig `yaml:"github"`
	DB DBConfig `yaml:"db"`
	Auth AuthConfig `yaml:"auth"`
}

func (c *Config) GetServerPort() uint16 {
	return c.Server.Port
}

func (c *Config) ServerAccessLogsAreEnabled() bool {
	return c.Server.AccessLogs.Enabled
}

func (c *Config) ServerAccessLogFormat() string {
	return c.Server.AccessLogs.Format
}

func (c *Config) LogFormat() string {
	return c.Logging.Format
}

func (c *Config) LogLevel() string {
	return c.Logging.Level
}

func (c *Config) GetGithubOauthClientId() string {
	return c.Github.Oauth.ClientId
}

func (c *Config) GetGithubOauthClientSecret() string {
	return c.Github.Oauth.ClientSecret
}


func (c *Config) GetAuthJWTPublicKeyPath() string {
	return c.Auth.JWT.PublicKeyFile
}

func (c *Config) GetAuthJWTPrivateKeyPath() string {
	return c.Auth.JWT.PrivateKeyFile
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			AccessLogs: ServerAccessLogsConfig{
				Format:  "json",
				Enabled: true,
			},
		},
		Github: GithubConfig{
			Oauth: GithubOauthConfig{
				ClientId:     "",
				ClientSecret: "",
			},
		},
		Logging: LoggingConfig{
			Level:  "error",
			Format: "json",
		},
		DB: DBConfig{
			Postgres: PostgresConfig{
				Host: "",
				Username: "",
				Password: "",
				Port: 5432,
				DBName: "panoptes",
				MaxConnections: 10,
				MinConnections: 3,
				SSLMode: PostgresSSLModeRequire,
			},
		},
		Auth: AuthConfig{
			JWT: JWTConfig{

			},
		},
	}
}
