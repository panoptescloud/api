package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/panoptescloud/api/pkg/config"
)

// Ensures we are using the sensible defaults we expect we are.
// Take a moment to sanity check if any of the defaults have changed as it could
// cause issues if this is running anywhere.
func Test_Defaults(t *testing.T) {
	cfg := config.Default()

	assert.Equal(t, uint16(8080), cfg.GetServerPort())
	assert.Equal(t, true, cfg.ServerAccessLogsAreEnabled())
	assert.Equal(t, "json", cfg.ServerAccessLogFormat())
	assert.Equal(t, "json", cfg.LogFormat())
	assert.Equal(t, "error", cfg.LogLevel())
	assert.Equal(t, "", cfg.GetGithubOauthClientId())
	assert.Equal(t, "", cfg.GetGithubOauthClientSecret())
}

// Ensures the getters are actually returning the relevant values from the config.
// All values here are purposely different from eachother and the default config,
// to ensure that the default config isn't working by some coincidence. E.g. the
// format for access logs and app logs both default to JSON, if one returned the
// value for the other, everything would look okay.
func Test_GettersAreWorking(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 9999,
			AccessLogs: config.ServerAccessLogsConfig{
				Format:  "text",
				Enabled: false,
			},
		},
		Logging: config.LoggingConfig{
			Level:  "warn",
			Format: "blah",
		},
		Github: config.GithubConfig{
			Oauth: config.GithubOauthConfig{
				ClientId:     "github_oauth_client_id",
				ClientSecret: "github_oauth_client_secret",
			},
		},
		DB: config.DBConfig{
			Postgres: config.PostgresConfig{
				Host:           "postgres_host",
				Username:       "postgres_username",
				Password:       "postgres_password",
				Port:           8746,
				DBName:         "postgres_db_name",
				MaxConnections: 17,
				MinConnections: 15,
				SSLMode:        config.PostgresSSLModeVerifyFull,
			},
		},
	}

	assert.Equal(t, uint16(9999), cfg.GetServerPort())
	assert.Equal(t, false, cfg.ServerAccessLogsAreEnabled())
	assert.Equal(t, "text", cfg.ServerAccessLogFormat())
	assert.Equal(t, "blah", cfg.LogFormat())
	assert.Equal(t, "warn", cfg.LogLevel())
	assert.Equal(t, "github_oauth_client_id", cfg.GetGithubOauthClientId())
	assert.Equal(t, "github_oauth_client_secret", cfg.GetGithubOauthClientSecret())
}
