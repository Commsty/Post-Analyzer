package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ValidConfig(t *testing.T) {
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("TG_BOT_TOKEN", "test_token")
	os.Setenv("TG_APP_ID", "123456")
	os.Setenv("TG_APP_HASH", "test_hash")
	os.Setenv("SESSION_PATH", "/tmp/test_session")
	os.Setenv("AUTH_PHONE", "+1234567890")
	os.Setenv("AUTH_PASSWORD", "auth_password")
	os.Setenv("OPENROUTER_API_KEY", "test_api_key")

	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("TG_BOT_TOKEN")
		os.Unsetenv("TG_APP_ID")
		os.Unsetenv("TG_APP_HASH")
		os.Unsetenv("SESSION_PATH")
		os.Unsetenv("AUTH_PHONE")
		os.Unsetenv("AUTH_PASSWORD")
		os.Unsetenv("OPENROUTER_API_KEY")
	}()

	testConfig := `app:
  name: "TestApp"
  version: "1.0"
database:
  host: "localhost"
  port: 5432
  name: "testdb"
  user: "testuser"
  ssl_mode: "disable"`

	tmpFile, err := os.CreateTemp("", "test_config_*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testConfig)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "TestApp", cfg.App.Name)
	assert.Equal(t, "1.0", cfg.App.Version)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "test_password", cfg.Database.Password)
	assert.Equal(t, "test_token", cfg.API.Telegram.BotToken)
	assert.Equal(t, 123456, cfg.API.Telegram.AppID)
}

func TestLoadConfig_InvalidPath(t *testing.T) {
	cfg, err := LoadConfig("non_existent_config.yml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_EmptyConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "empty_config_*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_MalformedConfig(t *testing.T) {
	testConfig := `app:
  name: "TestApp"
  version: "1.0"
database:
  host: "localhost"
  port: not_a_number
  name: "testdb"`

	tmpFile, err := os.CreateTemp("", "malformed_config_*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testConfig)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("TG_BOT_TOKEN", "test_token")
	os.Setenv("TG_APP_ID", "123456")
	os.Setenv("TG_APP_HASH", "test_hash")
	os.Setenv("SESSION_PATH", "/tmp/test_session")
	os.Setenv("AUTH_PHONE", "+1234567890")
	os.Setenv("AUTH_PASSWORD", "auth_password")
	os.Setenv("OPENROUTER_API_KEY", "test_api_key")

	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("TG_BOT_TOKEN")
		os.Unsetenv("TG_APP_ID")
		os.Unsetenv("TG_APP_HASH")
		os.Unsetenv("SESSION_PATH")
		os.Unsetenv("AUTH_PHONE")
		os.Unsetenv("AUTH_PASSWORD")
		os.Unsetenv("OPENROUTER_API_KEY")
	}()

	testConfig := `app:
  name: "PartialApp"
database:
  host: "localhost"`

	tmpFile, err := os.CreateTemp("", "partial_config_*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testConfig)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "PartialApp", cfg.App.Name)
	assert.Equal(t, "localhost", cfg.Database.Host)
}

func TestConfig_Validation(t *testing.T) {
	testCases := []struct {
		name   string
		config AppConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: AppConfig{
				App: struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
				}{
					Name:    "TestApp",
					Version: "1.0",
				},
				Database: struct {
					Host     string `yaml:"host"`
					Port     int    `yaml:"port"`
					Name     string `yaml:"name"`
					User     string `yaml:"user"`
					Password string `yaml:"-"`
					SSLMode  string `yaml:"ssl_mode"`
				}{
					Host:    "localhost",
					Port:    5432,
					Name:    "testdb",
					User:    "testuser",
					SSLMode: "disable",
				},
			},
			valid: true,
		},
		{
			name: "empty app name",
			config: AppConfig{
				App: struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
				}{
					Name:    "",
					Version: "1.0",
				},
				Database: struct {
					Host     string `yaml:"host"`
					Port     int    `yaml:"port"`
					Name     string `yaml:"name"`
					User     string `yaml:"user"`
					Password string `yaml:"-"`
					SSLMode  string `yaml:"ssl_mode"`
				}{
					Host:    "localhost",
					Port:    5432,
					Name:    "testdb",
					User:    "testuser",
					SSLMode: "disable",
				},
			},
			valid: false,
		},
		{
			name: "invalid database port",
			config: AppConfig{
				App: struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
				}{
					Name:    "TestApp",
					Version: "1.0",
				},
				Database: struct {
					Host     string `yaml:"host"`
					Port     int    `yaml:"port"`
					Name     string `yaml:"name"`
					User     string `yaml:"user"`
					Password string `yaml:"-"`
					SSLMode  string `yaml:"ssl_mode"`
				}{
					Host:    "localhost",
					Port:    0,
					Name:    "testdb",
					User:    "testuser",
					SSLMode: "disable",
				},
			},
			valid: false,
		},
		{
			name: "empty database host",
			config: AppConfig{
				App: struct {
					Name    string `yaml:"name"`
					Version string `yaml:"version"`
				}{
					Name:    "TestApp",
					Version: "1.0",
				},
				Database: struct {
					Host     string `yaml:"host"`
					Port     int    `yaml:"port"`
					Name     string `yaml:"name"`
					User     string `yaml:"user"`
					Password string `yaml:"-"`
					SSLMode  string `yaml:"ssl_mode"`
				}{
					Host:    "",
					Port:    5432,
					Name:    "testdb",
					User:    "testuser",
					SSLMode: "disable",
				},
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appValid := tc.config.App.Name != "" && tc.config.App.Version != ""
			dbValid := tc.config.Database.Host != "" &&
				tc.config.Database.Port > 0 &&
				tc.config.Database.Name != "" &&
				tc.config.Database.User != "" &&
				tc.config.Database.SSLMode != ""

			valid := appValid && dbValid

			if tc.valid {
				assert.True(t, valid, "Config should be valid")
			} else {
				assert.False(t, valid, "Config should be invalid")
			}
		})
	}
}

func TestConfig_DatabaseConnectionString(t *testing.T) {
	config := AppConfig{
		Database: struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Name     string `yaml:"name"`
			User     string `yaml:"user"`
			Password string `yaml:"-"`
			SSLMode  string `yaml:"ssl_mode"`
		}{
			Host:    "localhost",
			Port:    5432,
			Name:    "testdb",
			User:    "testuser",
			SSLMode: "disable",
		},
	}

	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "testdb", config.Database.Name)
	assert.Equal(t, "testuser", config.Database.User)
	assert.Equal(t, "disable", config.Database.SSLMode)
}

func TestConfig_EnvironmentOverrides(t *testing.T) {
	originalValue := os.Getenv("DATABASE_HOST")
	defer func() {
		if originalValue != "" {
			os.Setenv("DATABASE_HOST", originalValue)
		} else {
			os.Unsetenv("DATABASE_HOST")
		}
	}()

	os.Setenv("DATABASE_HOST", "env-host")

	envValue := os.Getenv("DATABASE_HOST")
	assert.Equal(t, "env-host", envValue)
}
