package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mkozhukh/youtrack/internal/mcp"
	"github.com/mkozhukh/youtrack/internal/mcp/logging"

	"github.com/charmbracelet/log"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

// Config holds the application configuration
type Config struct {
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
		DefaultProject string `koanf:"default_project"`
		Timeout        int    `koanf:"timeout"`
		DefaultQuery   string `koanf:"default_query"`
		MaxResults     int    `koanf:"max_results"`
	} `koanf:"youtrack"`
	Cache struct {
		TTLSeconds int `koanf:"ttl_seconds"`
	} `koanf:"cache"`
	Tracker struct {
		FilePath string `koanf:"file_path"`
	} `koanf:"tracker"`
}

var (
	k      = koanf.New(".")
	config Config
)

func main() {
	var useHTTP bool

	rootCmd := &cobra.Command{
		Use:   "youtrack-mcp",
		Short: "MCP server for YouTrack integration",
		Long:  `A Model Context Protocol (MCP) server that provides YouTrack integration capabilities.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, useHTTP)
		},
	}

	rootCmd.Flags().BoolVar(&useHTTP, "http", false, "Use StreamableHTTP transport instead of stdio")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() error {
	// Load defaults
	defaults := map[string]interface{}{
		"server.port":                 3204,
		"server.name":                 "YouTrack MCP Server",
		"logging.enabled":             false,
		"logging.call_log_path":       "calls.log",
		"logging.rest_error_log_path": "rest_errors.log",
		"logging.tool_error_log_path": "tool_errors.log",
		"youtrack.base_url":           "",
		"youtrack.api_key":            "",
		"youtrack.default_project":    "",
		"youtrack.timeout":            30,
		"youtrack.default_query":      "updated: -7d",
		"youtrack.max_results":        10,
		"cache.ttl_seconds":           300,
		"tracker.file_path":           "projects.json",
	}

	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return fmt.Errorf("error loading defaults: %w", err)
	}

	// Load from environment variables with YOUTRACK_ prefix
	// Environment variables are automatically mapped to config keys
	// e.g., YOUTRACK_YOUTRACK_BASE_URL -> youtrack.base_url
	if err := k.Load(env.Provider("YOUTRACK_", ".", func(s string) string {
		return s
	}), nil); err != nil {
		return fmt.Errorf("error loading env vars: %w", err)
	}

	// Load from config.toml if it exists
	configPath := "config.toml"
	if envPath := os.Getenv("YOUTRACK_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return fmt.Errorf("error loading config file: %w", err)
		}
		absPath, _ := filepath.Abs(configPath)
		log.Info("Loaded configuration from", "path", absPath)
	}

	// Unmarshal to struct
	if err := k.Unmarshal("", &config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

func run(cmd *cobra.Command, args []string, useHTTP bool) error {
	// Load configuration
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create server configuration
	serverConfig := mcp.ServerConfig{
		Name: config.Server.Name,
		Port: config.Server.Port,
		YouTrack: mcp.YouTrackConfig{
			BaseURL:        config.YouTrack.BaseURL,
			APIKey:         config.YouTrack.APIKey,
			DefaultProject: config.YouTrack.DefaultProject,
			Timeout:        config.YouTrack.Timeout,
			DefaultQuery:   config.YouTrack.DefaultQuery,
			MaxResults:     config.YouTrack.MaxResults,
		},
		Cache: mcp.CacheConfig{
			TTL: time.Duration(config.Cache.TTLSeconds) * time.Second,
		},
		Tracker: mcp.TrackerConfig{
			FilePath: config.Tracker.FilePath,
		},
		Logging: logging.LogConfig{
			Enabled:          config.Logging.Enabled,
			CallLogPath:      config.Logging.CallLogPath,
			RESTErrorLogPath: config.Logging.RESTErrorLogPath,
			ToolErrorLogPath: config.Logging.ToolErrorLogPath,
		},
		ToolBlacklist: config.Tools.Blacklist,
	}

	// Create a new MCP server with YouTrack integration
	s, err := mcp.NewMCPServer(serverConfig, logToolCall)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Register tools
	if err := s.RegisterTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Start the server with selected transport
	if useHTTP {
		log.Info("Starting server with StreamableHTTP transport")
		if err := s.ServeHTTP(); err != nil {
			return fmt.Errorf("HTTP server error: %w", err)
		}
	} else {
		log.Info("Starting server with stdio transport")
		if err := s.Serve(); err != nil {
			return fmt.Errorf("stdio server error: %w", err)
		}
	}

	return nil
}

func logToolCall(toolName string, args map[string]interface{}) {
	// Log tool call to console
	log.Info("Tool call", "tool", toolName, "args", args)
}
