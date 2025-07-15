package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// Config holds the application configuration
type Config struct {
	Server struct {
		Port int    `koanf:"port"`
		Name string `koanf:"name"`
	} `koanf:"server"`
	Logging struct {
		LogToolCalls  bool   `koanf:"log_tool_calls"`
		ToolCallsFile string `koanf:"tool_calls_file"`
	} `koanf:"logging"`
}

var (
	k             = koanf.New(".")
	config        Config
	toolCallsFile *os.File
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "youtrack-mcp",
		Short: "MCP server for YouTrack integration",
		Long:  `A Model Context Protocol (MCP) server that provides YouTrack integration capabilities.`,
		RunE:  run,
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() error {
	// Load defaults
	defaults := map[string]interface{}{
		"server.port":             3204,
		"server.name":             "YouTrack MCP Server",
		"logging.log_tool_calls":  false,
		"logging.tool_calls_file": "tool_calls.log",
	}

	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return fmt.Errorf("error loading defaults: %w", err)
	}

	// Load from environment variables with YOUTRACK_ prefix
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

func run(cmd *cobra.Command, args []string) error {
	// Load configuration
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log.Info("Starting YouTrack MCP Server", "name", config.Server.Name, "port", config.Server.Port)

	// Setup tool call logging if enabled
	if config.Logging.LogToolCalls {
		var err error
		toolCallsFile, err = os.OpenFile(config.Logging.ToolCallsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open tool calls log file: %w", err)
		}
		defer toolCallsFile.Close()
		log.Info("Tool call logging enabled", "file", config.Logging.ToolCallsFile)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		config.Server.Name,
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add tool handler
	s.AddTool(getHelloWorldTool(), helloHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func getHelloWorldTool() mcp.Tool {
	return mcp.NewTool("hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)
}

func logToolCall(toolName string, args map[string]interface{}) {
	// Log tool call to normal log
	log.Info("Tool call", "tool", toolName, "args", args)

	// Log to file if enabled
	if config.Logging.LogToolCalls && toolCallsFile != nil {
		logEntry := fmt.Sprintf("[%s] Tool: %s, Args: %v\n",
			time.Now().Format("2006-01-02 15:04:05"), toolName, args)
		toolCallsFile.WriteString(logEntry)
	}
}
