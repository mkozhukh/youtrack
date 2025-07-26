package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Defaults DefaultsConfig `koanf:"defaults"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	URL   string `koanf:"url"`
	Token string `koanf:"token"`
}

// DefaultsConfig holds default values
type DefaultsConfig struct {
	Project string `koanf:"project"`
	UserID  string `koanf:"user_id"`
}

// Global instance for the configuration
var k = koanf.New(".")

// Load loads configuration from file, environment variables, and CLI flags
func Load(configPath string, flags *pflag.FlagSet) (*Config, error) {
	var cfg Config

	// If no config path is provided, use the default
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".config", "yt", "config.toml")
	}

	// Load from config file if it exists
	if _, err := os.Stat(configPath); err == nil {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Load from environment variables with YT_ prefix
	if err := k.Load(env.Provider("YT_", ".", func(s string) string {
		// YT_SERVER_URL -> server.url
		// YT_SERVER_TOKEN -> server.token
		// YT_DEFAULTS_PROJECT -> defaults.project
		// YT_USER_ID -> defaults.user_id (special case for backward compatibility)
		s = strings.ToLower(strings.TrimPrefix(s, "YT_"))
		s = strings.ReplaceAll(s, "_", ".")

		// Special handling for YT_USER_ID -> defaults.user_id
		if s == "user.id" {
			return "defaults.user_id"
		}

		return s
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Load from command-line flags if provided
	if flags != nil {
		// Only load flags that were explicitly set
		if err := k.Load(posflag.ProviderWithFlag(flags, ".", k, func(f *pflag.Flag) (string, interface{}) {
			// Skip flags that weren't explicitly set
			if !f.Changed {
				return "", nil
			}

			// Map flag names to config keys
			switch f.Name {
			case "url":
				return "server.url", f.Value.String()
			case "token":
				return "server.token", f.Value.String()
			default:
				return "", nil
			}
		}), nil); err != nil {
			return nil, fmt.Errorf("failed to load command-line flags: %w", err)
		}
	}

	// Unmarshal into the config struct
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to the specified file
func Save(cfg *Config, configPath string) error {
	// If no config path is provided, use the default
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".config", "yt", "config.toml")
	}

	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal the config to TOML
	data, err := toml.Parser().Marshal(map[string]interface{}{
		"server": map[string]interface{}{
			"url":   cfg.Server.URL,
			"token": cfg.Server.Token,
		},
		"defaults": map[string]interface{}{
			"project": cfg.Defaults.Project,
			"user_id": cfg.Defaults.UserID,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "yt", "config.toml")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.URL == "" {
		return fmt.Errorf("server URL is required")
	}
	if c.Server.Token == "" {
		return fmt.Errorf("server token is required")
	}
	return nil
}
