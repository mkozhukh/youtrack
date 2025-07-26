package mcp

import (
	"fmt"
	"net/http"
	"time"

	"mkozhukh/youtrack/internal/mcp/handlers"
	"mkozhukh/youtrack/internal/mcp/tools"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/server"
)

// YouTrackConfig holds the YouTrack-specific configuration
type YouTrackConfig struct {
	BaseURL        string `koanf:"base_url"`
	APIKey         string `koanf:"api_key"`
	DefaultProject string `koanf:"default_project"`
	Timeout        int    `koanf:"timeout"`
	DefaultQuery   string `koanf:"default_query"`
	MaxResults     int    `koanf:"max_results"`
}

// ServerConfig holds the MCP server configuration
type ServerConfig struct {
	Name     string         `koanf:"name"`
	Port     int            `koanf:"port"`
	YouTrack YouTrackConfig `koanf:"youtrack"`
}

// MCPServer wraps the MCP server with YouTrack-specific functionality
type MCPServer struct {
	server          *server.MCPServer
	config          ServerConfig
	ytClient        *YouTrackClient
	toolLogger      func(string, map[string]interface{})
	issueHandlers   *handlers.IssueHandlers
	tagHandlers     *handlers.TagHandlers
	commentHandlers *handlers.CommentHandlers
	healthHandlers  *handlers.HealthHandlers
	startTime       time.Time
}

// NewMCPServer creates a new MCP server instance with YouTrack integration
func NewMCPServer(config ServerConfig, toolLogger func(string, map[string]interface{})) (*MCPServer, error) {
	// Create the underlying MCP server
	s := server.NewMCPServer(
		config.Name,
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Create YouTrack client
	ytClient, err := NewYouTrackClient(config.YouTrack)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTrack client: %w", err)
	}

	// Create issue handlers
	issueHandlers := handlers.NewIssueHandlers(ytClient, toolLogger)

	// Create tag handlers
	tagHandlers := handlers.NewTagHandlers(ytClient, toolLogger)

	// Create comment handlers
	commentHandlers := handlers.NewCommentHandlers(ytClient, toolLogger)

	// Create health handlers
	startTime := time.Now()
	healthHandlers := handlers.NewHealthHandlers(ytClient, toolLogger, startTime)

	return &MCPServer{
		server:          s,
		config:          config,
		ytClient:        ytClient,
		toolLogger:      toolLogger,
		issueHandlers:   issueHandlers,
		tagHandlers:     tagHandlers,
		commentHandlers: commentHandlers,
		healthHandlers:  healthHandlers,
		startTime:       startTime,
	}, nil
}

// RegisterTools registers all YouTrack-related tools with the MCP server
func (s *MCPServer) RegisterTools() error {
	// Register issue management tools
	s.server.AddTool(tools.GetIssueListTool(), s.issueHandlers.GetIssueListHandler)
	s.server.AddTool(tools.GetIssueDetailsTool(), s.issueHandlers.GetIssueDetailsHandler)
	s.server.AddTool(tools.CreateIssueTool(), s.issueHandlers.CreateIssueHandler)
	s.server.AddTool(tools.UpdateIssueTool(), s.issueHandlers.UpdateIssueHandler)

	// Register tag management tools
	s.server.AddTool(tools.TagIssueTool(), s.tagHandlers.TagIssueHandler)

	// Register comment management tools
	s.server.AddTool(tools.AddCommentTool(), s.commentHandlers.AddCommentHandler)

	return nil
}

// Serve starts the MCP server using stdio transport
func (s *MCPServer) Serve() error {
	return server.ServeStdio(s.server)
}

// ServeHTTP starts the MCP server using StreamableHTTP transport
func (s *MCPServer) ServeHTTP() error {
	// Create StreamableHTTP server
	streamableServer := server.NewStreamableHTTPServer(s.server)

	// StreamableHTTPServer implements http.Handler, so we can use it directly
	http.Handle("/mcp", streamableServer)

	// Add health endpoint
	http.HandleFunc("/health", s.healthHandlers.HealthCheckHTTPHandler)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Info("Starting StreamableHTTP server", "address", addr)

	return http.ListenAndServe(addr, nil)
}

// GetYouTrackClient returns the YouTrack client for use in handlers
func (s *MCPServer) GetYouTrackClient() *YouTrackClient {
	return s.ytClient
}
