package config

type ServerAccessLogsConfig struct {
	Format string
	Enabled bool
}

type ServerConfig struct {
	Port uint16
	AccessLogs ServerAccessLogsConfig `yaml:"accessLogs"`
}

type GithubOauthConfig struct {
	ClientId string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

type GithubConfig struct {
	Oauth GithubOauthConfig `yaml:"oauth"`
}

type LoggingConfig struct {
	Level string
	Format string
}

type Config struct {
	Server ServerConfig
	Logging LoggingConfig
	Github GithubConfig `yaml:"github"`
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

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			AccessLogs: ServerAccessLogsConfig{
				Format: "json",
				Enabled: true,
			},
		},
		Github: GithubConfig{
			Oauth: GithubOauthConfig{
				ClientId: "",
				ClientSecret: "",
			},
		},
		Logging: LoggingConfig{
			Level: "error",
			Format: "json",
		},
	}
}