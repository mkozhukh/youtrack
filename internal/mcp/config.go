package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mkozhukh/youtrack/internal/mcp/logging"

	"github.com/charmbracelet/log"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// fileConfig mirrors the TOML file structure for loading.
type fileConfig struct {
	Server struct {
		Port int    `koanf:"port"`
		Name string `koanf:"name"`
	} `koanf:"server"`
	Logging struct {
		Enabled          bool   `koanf:"enabled"`
		CallLogPath      string `koanf:"call_log_path"`
		RESTErrorLogPath string `koanf:"rest_error_log_path"`
		ToolErrorLogPath string `koanf:"tool_error_log_path"`
	} `koanf:"logging"`
	Tools struct {
		Blacklist []string `koanf:"blacklist"`
	} `koanf:"tools"`
	YouTrack struct {
		BaseURL        string `koanf:"base_url"`
		APIKey         string `koanf:"api_key"`
		HubURL         string `koanf:"hub_url"`
		DefaultProject string `koanf:"default_project"`
		Timeout        int    `koanf:"timeout"`
		MaxResults     int    `koanf:"max_results"`
	} `koanf:"youtrack"`
	Cache struct {
		TTLSeconds int `koanf:"ttl_seconds"`
	} `koanf:"cache"`
	Tracker struct {
		FilePath string `koanf:"file_path"`
	} `koanf:"tracker"`
	FileServer struct {
		Enabled       bool   `koanf:"enabled"`
		BaseURL       string `koanf:"base_url"`
		TTLSeconds    int    `koanf:"ttl_seconds"`
		MaxFileSizeMB int    `koanf:"max_file_size_mb"`
	} `koanf:"fileserver"`
}

// LoadConfig loads ServerConfig from a TOML file and environment variables.
// configPath defaults to "config.toml"; override via YOUTRACK_CONFIG_PATH env var.
func LoadConfig(configPath string) (ServerConfig, error) {
	k := koanf.New(".")

	defaults := map[string]any{
		"server.port":                 3204,
		"server.name":                 "YouTrack MCP Server",
		"logging.enabled":             false,
		"logging.call_log_path":       "calls.log",
		"logging.rest_error_log_path": "rest_errors.log",
		"logging.tool_error_log_path": "tool_errors.log",
		"youtrack.base_url":           "",
		"youtrack.api_key":            "",
		"youtrack.hub_url":            "",
		"youtrack.default_project":    "",
		"youtrack.timeout":            30,
		"youtrack.max_results":        10,
		"cache.ttl_seconds":           300,
		"tracker.file_path":           "projects.json",
		"fileserver.enabled":          false,
		"fileserver.base_url":         "",
		"fileserver.ttl_seconds":      1800,
		"fileserver.max_file_size_mb": 50,
	}

	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return ServerConfig{}, fmt.Errorf("error loading defaults: %w", err)
	}

	if err := k.Load(env.Provider("YOUTRACK_", ".", func(s string) string {
		return s
	}), nil); err != nil {
		return ServerConfig{}, fmt.Errorf("error loading env vars: %w", err)
	}

	if configPath == "" {
		configPath = "config.toml"
	}
	if envPath := os.Getenv("YOUTRACK_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return ServerConfig{}, fmt.Errorf("error loading config file: %w", err)
		}
		absPath, _ := filepath.Abs(configPath)
		log.Info("Loaded configuration from", "path", absPath)
	}

	var fc fileConfig
	if err := k.Unmarshal("", &fc); err != nil {
		return ServerConfig{}, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return ServerConfig{
		Name: fc.Server.Name,
		Port: fc.Server.Port,
		YouTrack: YouTrackConfig{
			BaseURL:        fc.YouTrack.BaseURL,
			APIKey:         fc.YouTrack.APIKey,
			HubURL:         fc.YouTrack.HubURL,
			DefaultProject: fc.YouTrack.DefaultProject,
			Timeout:        fc.YouTrack.Timeout,
			MaxResults:     fc.YouTrack.MaxResults,
		},
		Cache: CacheConfig{
			TTL: time.Duration(fc.Cache.TTLSeconds) * time.Second,
		},
		Tracker: TrackerConfig{
			FilePath: fc.Tracker.FilePath,
		},
		FileServer: FileServerConfig{
			Enabled:       fc.FileServer.Enabled,
			BaseURL:       fc.FileServer.BaseURL,
			TTLSeconds:    fc.FileServer.TTLSeconds,
			MaxFileSizeMB: fc.FileServer.MaxFileSizeMB,
		},
		Logging: logging.LogConfig{
			Enabled:          fc.Logging.Enabled,
			CallLogPath:      fc.Logging.CallLogPath,
			RESTErrorLogPath: fc.Logging.RESTErrorLogPath,
			ToolErrorLogPath: fc.Logging.ToolErrorLogPath,
		},
		ToolBlacklist: fc.Tools.Blacklist,
	}, nil
}
