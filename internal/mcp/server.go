package mcp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mkozhukh/youtrack/internal/mcp/cache"
	"github.com/mkozhukh/youtrack/internal/mcp/handlers"
	"github.com/mkozhukh/youtrack/internal/mcp/logging"
	"github.com/mkozhukh/youtrack/internal/mcp/tools"
	"github.com/mkozhukh/youtrack/internal/mcp/tracker"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
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

// CacheConfig holds cache-specific configuration
type CacheConfig struct {
	TTL time.Duration
}

// TrackerConfig holds project tracker configuration
type TrackerConfig struct {
	FilePath string
}

// ServerConfig holds the MCP server configuration
type ServerConfig struct {
	Name          string         `koanf:"name"`
	Port          int            `koanf:"port"`
	YouTrack      YouTrackConfig `koanf:"youtrack"`
	Cache         CacheConfig
	Tracker       TrackerConfig
	Logging       logging.LogConfig
	ToolBlacklist []string
}

// MCPServer wraps the MCP server with YouTrack-specific functionality
type MCPServer struct {
	server             *server.MCPServer
	config             ServerConfig
	ytClient           *YouTrackClient
	cachedClient       *cache.CachedClient
	appLogger          *logging.AppLogger
	toolLogger         func(string, map[string]interface{})
	issueHandlers      *handlers.IssueHandlers
	tagHandlers        *handlers.TagHandlers
	commentHandlers    *handlers.CommentHandlers
	healthHandlers     *handlers.HealthHandlers
	projectHandlers    *handlers.ProjectHandlers
	userHandlers       *handlers.UserHandlers
	linkHandlers       *handlers.LinkHandlers
	attachmentHandlers *handlers.AttachmentHandlers
	commandHandlers    *handlers.CommandHandlers
	worklogHandlers    *handlers.WorklogHandlers
	cacheHandlers      *handlers.CacheHandlers
	startTime          time.Time
}

// NewMCPServer creates a new MCP server instance with YouTrack integration
func NewMCPServer(config ServerConfig, toolLogger func(string, map[string]interface{})) (*MCPServer, error) {
	// Create the underlying MCP server
	s := server.NewMCPServer(
		config.Name,
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Create app logger
	var appLogger *logging.AppLogger
	if config.Logging.Enabled {
		var err error
		appLogger, err = logging.NewAppLogger(config.Logging)
		if err != nil {
			return nil, fmt.Errorf("failed to create app logger: %w", err)
		}
		log.Info("Structured logging enabled",
			"call_log", config.Logging.CallLogPath,
			"rest_error_log", config.Logging.RESTErrorLogPath,
			"tool_error_log", config.Logging.ToolErrorLogPath)
	}

	// Wrap the toolLogger to also log to the app logger
	wrappedToolLogger := func(toolName string, args map[string]interface{}) {
		// Call the original tool logger
		if toolLogger != nil {
			toolLogger(toolName, args)
		}
		// Also log to the app logger (keyHash will be empty for now - can be enhanced for HTTP mode)
		if appLogger != nil {
			keyHash := logging.HashAPIKey(config.YouTrack.APIKey)
			appLogger.LogToolCall(keyHash, toolName)
		}
	}

	// Create YouTrack client
	ytClient, err := NewYouTrackClient(config.YouTrack, appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTrack client: %w", err)
	}

	// Create cache with configured TTL (default to 5 minutes if not set)
	cacheTTL := config.Cache.TTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}
	projectCache := cache.NewProjectCache(cacheTTL)
	cachedClient := cache.NewCachedClient(ytClient, projectCache)

	log.Info("Cache initialized", "ttl", cacheTTL)

	// Create project tracker
	projectTracker := tracker.NewProjectTracker(config.Tracker.FilePath)
	contextTracker := tracker.NewContextProjectTracker(projectTracker, ytClient)

	if config.Tracker.FilePath != "" {
		log.Info("Project tracker initialized", "file", config.Tracker.FilePath)
	}

	// Create issue handlers
	issueHandlers := handlers.NewIssueHandlers(ytClient, wrappedToolLogger, contextTracker)

	// Create tag handlers
	tagHandlers := handlers.NewTagHandlers(ytClient, wrappedToolLogger)

	// Create comment handlers
	commentHandlers := handlers.NewCommentHandlers(ytClient, wrappedToolLogger)

	// Create health handlers
	startTime := time.Now()
	healthHandlers := handlers.NewHealthHandlers(ytClient, wrappedToolLogger, startTime)

	// Create project handlers with cached client
	projectHandlers := handlers.NewProjectHandlers(cachedClient, wrappedToolLogger, contextTracker)

	// Create user handlers with cached client
	userHandlers := handlers.NewUserHandlers(cachedClient, config.YouTrack.DefaultProject, wrappedToolLogger, contextTracker)

	// Create link handlers
	linkHandlers := handlers.NewLinkHandlers(ytClient, wrappedToolLogger)

	// Create attachment handlers
	attachmentHandlers := handlers.NewAttachmentHandlers(ytClient, wrappedToolLogger)

	// Create command handlers
	commandHandlers := handlers.NewCommandHandlers(ytClient, wrappedToolLogger)

	// Create worklog handlers
	worklogHandlers := handlers.NewWorklogHandlers(ytClient, wrappedToolLogger)

	// Create cache handlers
	cacheHandlers := handlers.NewCacheHandlers(projectCache, wrappedToolLogger)

	return &MCPServer{
		server:             s,
		config:             config,
		ytClient:           ytClient,
		cachedClient:       cachedClient,
		appLogger:          appLogger,
		toolLogger:         toolLogger,
		issueHandlers:      issueHandlers,
		tagHandlers:        tagHandlers,
		commentHandlers:    commentHandlers,
		healthHandlers:     healthHandlers,
		projectHandlers:    projectHandlers,
		userHandlers:       userHandlers,
		linkHandlers:       linkHandlers,
		attachmentHandlers: attachmentHandlers,
		commandHandlers:    commandHandlers,
		worklogHandlers:    worklogHandlers,
		cacheHandlers:      cacheHandlers,
		startTime:          startTime,
	}, nil
}

// GetAppLogger returns the app logger
func (s *MCPServer) GetAppLogger() *logging.AppLogger {
	return s.appLogger
}

// isBlacklisted checks if a tool name is in the blacklist
func (s *MCPServer) isBlacklisted(toolName string) bool {
	for _, name := range s.config.ToolBlacklist {
		if name == toolName {
			return true
		}
	}
	return false
}

// addTool registers a tool if it's not blacklisted
func (s *MCPServer) addTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	if s.isBlacklisted(tool.Name) {
		log.Info("Tool blacklisted, skipping", "tool", tool.Name)
		return
	}
	s.server.AddTool(tool, handler)
}

// RegisterTools registers all YouTrack-related tools with the MCP server
func (s *MCPServer) RegisterTools() error {
	// Register issue management tools
	s.addTool(tools.GetIssueListTool(), s.issueHandlers.GetIssueListHandler)
	s.addTool(tools.GetIssueDetailsTool(), s.issueHandlers.GetIssueDetailsHandler)
	s.addTool(tools.CreateIssueTool(), s.issueHandlers.CreateIssueHandler)
	s.addTool(tools.UpdateIssueTool(), s.issueHandlers.UpdateIssueHandler)
	s.addTool(tools.DeleteIssueTool(), s.issueHandlers.DeleteIssueHandler)

	// Register tag management tools
	s.addTool(tools.TagIssueTool(), s.tagHandlers.TagIssueHandler)
	s.addTool(tools.UntagIssueTool(), s.tagHandlers.UntagIssueHandler)
	s.addTool(tools.SearchTagsTool(), s.tagHandlers.SearchTagsHandler)

	// Register comment management tools
	s.addTool(tools.AddCommentTool(), s.commentHandlers.AddCommentHandler)

	// Register project management tools
	s.addTool(tools.GetProjectInfoTool(), s.projectHandlers.GetProjectInfoHandler)
	s.addTool(tools.ListProjectsTool(), s.projectHandlers.ListProjectsHandler)

	// Register user management tools
	s.addTool(tools.GetCurrentUserTool(), s.userHandlers.GetCurrentUserHandler)
	s.addTool(tools.GetProjectUsersTool(), s.userHandlers.GetProjectUsersHandler)

	// Register link management tools
	s.addTool(tools.GetIssueLinksTool(), s.linkHandlers.GetIssueLinksHandler)
	s.addTool(tools.CreateIssueLinkTool(), s.linkHandlers.CreateIssueLinkHandler)

	// Register attachment management tools
	s.addTool(tools.GetIssueAttachmentsTool(), s.attachmentHandlers.GetIssueAttachmentsHandler)
	s.addTool(tools.GetIssueAttachmentContentTool(), s.attachmentHandlers.GetIssueAttachmentContentHandler)
	s.addTool(tools.UploadAttachmentTool(), s.attachmentHandlers.UploadAttachmentHandler)

	// Register command tools
	s.addTool(tools.ApplyCommandTool(), s.commandHandlers.ApplyCommandHandler)

	// Register worklog tools
	s.addTool(tools.AddWorklogTool(), s.worklogHandlers.AddWorklogHandler)
	s.addTool(tools.GetIssueWorklogsTool(), s.worklogHandlers.GetIssueWorklogsHandler)
	s.addTool(tools.GetUserWorklogsTool(), s.worklogHandlers.GetUserWorklogsHandler)

	// Register cache management tools
	s.addTool(tools.DropCacheTool(), s.cacheHandlers.DropCacheHandler)

	return nil
}

// Serve starts the MCP server using stdio transport
func (s *MCPServer) Serve() error {
	return server.ServeStdio(s.server)
}

// CORSMiddleware adds CORS headers for browser-based clients
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Cache-Control")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ServeHTTP starts the MCP server using StreamableHTTP transport
func (s *MCPServer) ServeHTTP() error {
	// Create StreamableHTTP server
	streamableServer := server.NewStreamableHTTPServer(s.server)

	// Wrap with CORS and auth middleware
	http.Handle("/mcp", CORSMiddleware(AuthMiddleware(streamableServer)))

	// Add health endpoint
	http.HandleFunc("/health", s.healthHandlers.HealthCheckHTTPHandler)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Info("Starting StreamableHTTP server", "address", addr)

	if s.config.YouTrack.APIKey == "" {
		log.Info("Per-request auth mode: clients must provide Authorization header")
	}

	return http.ListenAndServe(addr, nil)
}

// GetYouTrackClient returns the YouTrack client for use in handlers
func (s *MCPServer) GetYouTrackClient() *YouTrackClient {
	return s.ytClient
}
