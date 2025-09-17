package config

type ServerAccessLogsConfig struct {
	Format string
	Enabled bool
}

type ServerConfig struct {
	Port int
	AccessLogs ServerAccessLogsConfig `yaml:"accessLogs" mapstructure:"access_logs"`
}

type LoggingConfig struct {
	Level string
	Format string
}

type Config struct {
	Server ServerConfig
	Logging LoggingConfig
}

func (c *Config) GetServerPort() int {
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

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			AccessLogs: ServerAccessLogsConfig{
				Format: "json",
				Enabled: true,
			},
		},
		Logging: LoggingConfig{
			Level: "error",
			Format: "json",
		},
	}
}