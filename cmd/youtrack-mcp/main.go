package main

import (
	"fmt"

	"github.com/mkozhukh/youtrack/internal/mcp"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func main() {
	var useHTTP bool

	rootCmd := &cobra.Command{
		Use:   "youtrack-mcp",
		Short: "MCP server for YouTrack integration",
		Long:  `A Model Context Protocol (MCP) server that provides YouTrack integration capabilities.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(useHTTP)
		},
	}

	rootCmd.Flags().BoolVar(&useHTTP, "http", false, "Use StreamableHTTP transport instead of stdio")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(useHTTP bool) error {
	serverConfig, err := mcp.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	s, err := mcp.NewMCPServer(serverConfig, logToolCall)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	if err := s.RegisterTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

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

func logToolCall(toolName string, args map[string]any) {
	log.Info("Tool call", "tool", toolName, "args", args)
}
