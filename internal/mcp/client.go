package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkozhukh/youtrack/internal/mcp/logging"
	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/charmbracelet/log"
)

// YouTrackClient wraps the YouTrack client with configuration and context handling
type YouTrackClient struct {
	client     *youtrack.Client
	config     YouTrackConfig
	defaultCtx *youtrack.YouTrackContext
	appLogger  *logging.AppLogger
}

// NewYouTrackClient creates a new YouTrack client wrapper with configuration
func NewYouTrackClient(config YouTrackConfig, appLogger *logging.AppLogger) (*YouTrackClient, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create the base client
	client := youtrack.NewClient(config.BaseURL)

	// Set up REST logger if app logger is provided and API key is available
	if appLogger != nil && config.APIKey != "" {
		keyHash := logging.HashAPIKey(config.APIKey)
		client.SetLogger(appLogger.NewRESTLoggerWithContext(keyHash))
	}

	// Note: The client timeout is set in the NewClient constructor
	// We would need to modify the client constructor to support custom timeouts

	// Create default context
	defaultCtx := youtrack.NewYouTrackContext(context.Background(), config.APIKey)

	ytClient := &YouTrackClient{
		client:     client,
		config:     config,
		defaultCtx: defaultCtx,
		appLogger:  appLogger,
	}

	// Test connection only if API key is configured
	// When using per-request auth, connection will be tested on first request
	if config.APIKey != "" {
		if err := ytClient.testConnection(); err != nil {
			return nil, fmt.Errorf("failed to connect to YouTrack: %w", err)
		}
		log.Info("YouTrack client initialized successfully", "base_url", config.BaseURL)
	} else {
		log.Info("YouTrack client initialized (per-request auth mode)", "base_url", config.BaseURL)
	}

	return ytClient, nil
}

// validateConfig validates the YouTrack configuration
func validateConfig(config YouTrackConfig) error {
	if config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	// Validate URL format
	if !strings.HasPrefix(config.BaseURL, "http://") && !strings.HasPrefix(config.BaseURL, "https://") {
		return fmt.Errorf("base_url must start with http:// or https://")
	}

	// API key is optional - can be provided via HTTP header instead
	// If provided, validate format
	if config.APIKey != "" && len(config.APIKey) < 10 {
		return fmt.Errorf("api_key appears to be too short (minimum 10 characters)")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative")
	}

	if config.Timeout > 300 {
		return fmt.Errorf("timeout must be less than or equal to 300 seconds")
	}

	if config.MaxResults < 0 {
		return fmt.Errorf("max_results must be non-negative")
	}

	if config.MaxResults > 1000 {
		return fmt.Errorf("max_results must be less than or equal to 1000")
	}

	// Validate default project if provided
	if config.DefaultProject != "" {
		if len(config.DefaultProject) < 2 {
			return fmt.Errorf("default_project must be at least 2 characters")
		}
		if len(config.DefaultProject) > 50 {
			return fmt.Errorf("default_project must be less than 50 characters")
		}
	}

	return nil
}

// testConnection tests the YouTrack connection by making a simple API call
func (c *YouTrackClient) testConnection() error {
	log.Info("Testing YouTrack connectivity", "base_url", c.config.BaseURL)

	// Try to search for projects to test the connection
	projects, err := c.client.ListProjects(c.defaultCtx, 0, 1)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Additional validation - check if we can access projects
	if projects == nil {
		return fmt.Errorf("connection test failed: received nil projects response")
	}

	log.Info("YouTrack connectivity test successful", "projects_accessible", len(projects))

	// Test default project access if configured
	if c.config.DefaultProject != "" {
		log.Info("Testing default project access", "project", c.config.DefaultProject)
		project, err := c.client.GetProject(c.defaultCtx, c.config.DefaultProject)
		if err != nil {
			return fmt.Errorf("default project '%s' is not accessible: %w", c.config.DefaultProject, err)
		}
		if project == nil {
			return fmt.Errorf("default project '%s' not found", c.config.DefaultProject)
		}
		log.Info("Default project access confirmed", "project", project.Name, "id", project.ID)
	}

	return nil
}

// GetConfig returns the YouTrack configuration
func (c *YouTrackClient) GetConfig() YouTrackConfig {
	return c.config
}

// WithContext creates a new context for API calls
// It checks for an auth token in the context first (from HTTP header),
// then falls back to the configured API key
func (c *YouTrackClient) WithContext(ctx context.Context) *youtrack.YouTrackContext {
	// Check for per-request auth token in context (from HTTP header)
	if token := GetAuthToken(ctx); token != "" {
		return youtrack.NewYouTrackContext(ctx, token)
	}

	// Fall back to configured API key
	return youtrack.NewYouTrackContext(ctx, c.config.APIKey)
}

// GetEffectiveAPIKey returns the API key that would be used for a given context
// Returns empty string if no key is available
func (c *YouTrackClient) GetEffectiveAPIKey(ctx context.Context) string {
	if token := GetAuthToken(ctx); token != "" {
		return token
	}
	return c.config.APIKey
}

// HasValidAuth checks if there's a valid auth token available (from context or config)
func (c *YouTrackClient) HasValidAuth(ctx context.Context) bool {
	return c.GetEffectiveAPIKey(ctx) != ""
}

// GetDefaultContext returns the default context for API calls
func (c *YouTrackClient) GetDefaultContext() *youtrack.YouTrackContext {
	return c.defaultCtx
}

// GetClient returns the underlying YouTrack client
func (c *YouTrackClient) GetClient() *youtrack.Client {
	return c.client
}

// Issue Management Methods

// GetIssue retrieves an issue by ID
func (c *YouTrackClient) GetIssue(ctx context.Context, issueID string) (*youtrack.Issue, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetIssue(ytCtx, issueID)
}

// SearchIssues searches for issues with optional parameters
func (c *YouTrackClient) SearchIssues(ctx context.Context, query string, skip, top int) ([]*youtrack.Issue, error) {
	ytCtx := c.WithContext(ctx)

	// Use default query if empty
	if query == "" {
		query = c.config.DefaultQuery
	}

	// Use default max results if top is 0
	if top == 0 {
		top = c.config.MaxResults
	}

	return c.client.SearchIssues(ytCtx, query, skip, top)
}

// CreateIssue creates a new issue
func (c *YouTrackClient) CreateIssue(ctx context.Context, req *youtrack.CreateIssueRequest) (*youtrack.Issue, error) {
	ytCtx := c.WithContext(ctx)

	// Use default project if not specified
	if req.Project.ID == "" && c.config.DefaultProject != "" {
		req.Project = youtrack.ProjectRef{ID: c.config.DefaultProject}
	}

	return c.client.CreateIssue(ytCtx, req)
}

// UpdateIssue updates an existing issue
func (c *YouTrackClient) UpdateIssue(ctx context.Context, issueID string, req *youtrack.UpdateIssueRequest) (*youtrack.Issue, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.UpdateIssue(ytCtx, issueID, req)
}

// UpdateIssueAssignee updates an issue's assignee
func (c *YouTrackClient) UpdateIssueAssignee(ctx context.Context, issueID string, assigneeLogin string) (*youtrack.Issue, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.UpdateIssueAssignee(ytCtx, issueID, assigneeLogin)
}

// UpdateIssueAssigneeByProject updates an issue's assignee by project
func (c *YouTrackClient) UpdateIssueAssigneeByProject(ctx context.Context, issueID string, projectID string, username string) (*youtrack.Issue, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.UpdateIssueAssigneeByProject(ytCtx, issueID, projectID, username)
}

// DeleteIssue deletes an issue
func (c *YouTrackClient) DeleteIssue(ctx context.Context, issueID string) error {
	ytCtx := c.WithContext(ctx)
	return c.client.DeleteIssue(ytCtx, issueID)
}

// Comment Management Methods

// GetIssueComments retrieves comments for an issue
func (c *YouTrackClient) GetIssueComments(ctx context.Context, issueID string) ([]*youtrack.IssueComment, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetIssueComments(ytCtx, issueID)
}

// AddIssueComment adds a comment to an issue
func (c *YouTrackClient) AddIssueComment(ctx context.Context, issueID string, comment string) (*youtrack.IssueComment, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.AddIssueComment(ytCtx, issueID, comment)
}

// Tag Management Methods

// AddIssueTag adds a tag to an issue by tag ID
func (c *YouTrackClient) AddIssueTag(ctx context.Context, issueID string, tagID string) error {
	ytCtx := c.WithContext(ctx)
	return c.client.AddIssueTag(ytCtx, issueID, tagID)
}

// EnsureTag ensures a tag exists, returns the tag ID
func (c *YouTrackClient) EnsureTag(ctx context.Context, tagName string, color string) (string, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.EnsureTag(ytCtx, tagName, color)
}

// User Management Methods

// GetUser returns a user by ID
func (c *YouTrackClient) GetUser(ctx context.Context, userID string) (*youtrack.User, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetUser(ytCtx, userID)
}

// GetUserByLogin returns a user by login
func (c *YouTrackClient) GetUserByLogin(ctx context.Context, login string) (*youtrack.User, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetUserByLogin(ytCtx, login)
}

// SuggestUserByProject suggests a user by project
func (c *YouTrackClient) SuggestUserByProject(ctx context.Context, projectID string, username string) (*youtrack.User, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.SuggestUserByProject(ytCtx, projectID, username)
}

// Project Management Methods

// GetProject returns a project by ID
func (c *YouTrackClient) GetProject(ctx context.Context, projectID string) (*youtrack.Project, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetProject(ytCtx, projectID)
}

// ListProjects returns all projects
func (c *YouTrackClient) ListProjects(ctx context.Context, skip, top int) ([]*youtrack.Project, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.ListProjects(ytCtx, skip, top)
}

// GetProjectByName returns a project by name
func (c *YouTrackClient) GetProjectByName(ctx context.Context, name string) (*youtrack.Project, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetProjectByName(ytCtx, name)
}

// GetProjectCustomFields returns the custom fields for a project
func (c *YouTrackClient) GetProjectCustomFields(ctx context.Context, projectID string) ([]*youtrack.CustomField, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetProjectCustomFields(ytCtx, projectID)
}

// GetCustomFieldAllowedValues returns the allowed values for a custom field in a project
func (c *YouTrackClient) GetCustomFieldAllowedValues(ctx context.Context, projectID string, fieldName string) ([]youtrack.AllowedValue, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetCustomFieldAllowedValues(ytCtx, projectID, fieldName)
}

// GetAvailableLinkTypes returns all available link types
func (c *YouTrackClient) GetAvailableLinkTypes(ctx context.Context) ([]*youtrack.LinkType, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetAvailableLinkTypes(ytCtx)
}

// GetIssueCustomFields returns the custom field values for an issue
func (c *YouTrackClient) GetIssueCustomFields(ctx context.Context, issueID string) ([]*youtrack.CustomFieldValue, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetIssueCustomFields(ytCtx, issueID)
}

// ApplyCommand applies a command to an issue
func (c *YouTrackClient) ApplyCommand(ctx context.Context, issueID string, command string) error {
	ytCtx := c.WithContext(ctx)
	return c.client.ApplyCommand(ytCtx, issueID, command)
}

// SearchIssuesSorted searches for issues with sorting
func (c *YouTrackClient) SearchIssuesSorted(ctx context.Context, query string, skip, top int, sortBy, sortOrder string) ([]*youtrack.Issue, error) {
	ytCtx := c.WithContext(ctx)

	// Use default query if empty
	if query == "" {
		query = c.config.DefaultQuery
	}

	// Use default max results if top is 0
	if top == 0 {
		top = c.config.MaxResults
	}

	return c.client.SearchIssuesSorted(ytCtx, query, skip, top, sortBy, sortOrder)
}

// GetIssueLinks returns the links for an issue
func (c *YouTrackClient) GetIssueLinks(ctx context.Context, issueID string) ([]*youtrack.IssueLink, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetIssueLinks(ytCtx, issueID)
}

// CreateIssueLink creates a link between two issues
func (c *YouTrackClient) CreateIssueLink(ctx context.Context, sourceID, targetID, linkType string) error {
	ytCtx := c.WithContext(ctx)
	return c.client.CreateIssueLink(ytCtx, sourceID, targetID, linkType)
}

// GetIssueAttachments returns the attachments for an issue
func (c *YouTrackClient) GetIssueAttachments(ctx context.Context, issueID string) ([]*youtrack.Attachment, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetIssueAttachments(ytCtx, issueID)
}

// GetIssueAttachmentContent downloads the content of an attachment
func (c *YouTrackClient) GetIssueAttachmentContent(ctx context.Context, issueID string, attachmentID string) ([]byte, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetIssueAttachmentContent(ytCtx, issueID, attachmentID)
}

// DownloadByURL downloads raw content from a YouTrack URL
func (c *YouTrackClient) DownloadByURL(ctx context.Context, rawURL string) ([]byte, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.DownloadByURL(ytCtx, rawURL)
}

// ListTags returns all tags
func (c *YouTrackClient) ListTags(ctx context.Context, skip, top int) ([]*youtrack.Tag, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.ListTags(ytCtx, skip, top)
}

// GetCurrentUser returns the currently authenticated user
func (c *YouTrackClient) GetCurrentUser(ctx context.Context) (*youtrack.User, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetCurrentUser(ytCtx)
}

// RemoveIssueTag removes a tag from an issue by tag ID
func (c *YouTrackClient) RemoveIssueTag(ctx context.Context, issueID string, tagID string) error {
	ytCtx := c.WithContext(ctx)
	return c.client.RemoveIssueTag(ytCtx, issueID, tagID)
}

// GetTagByName returns a tag by name
func (c *YouTrackClient) GetTagByName(ctx context.Context, name string) (*youtrack.Tag, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetTagByName(ytCtx, name)
}

// GetProjectUsers returns users in a project
func (c *YouTrackClient) GetProjectUsers(ctx context.Context, projectID string, skip, top int) ([]*youtrack.User, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetProjectUsers(ytCtx, projectID, skip, top)
}

// AddIssueAttachmentFromBytes uploads content as an attachment to an issue
func (c *YouTrackClient) AddIssueAttachmentFromBytes(ctx context.Context, issueID string, content []byte, filename string) (*youtrack.Attachment, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.AddIssueAttachmentFromBytes(ytCtx, issueID, content, filename)
}

// GetIssueWorklogs returns worklogs for an issue
func (c *YouTrackClient) GetIssueWorklogs(ctx context.Context, issueID string) ([]*youtrack.WorkItem, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetIssueWorklogs(ytCtx, issueID)
}

// AddIssueWorklog adds a worklog to an issue
func (c *YouTrackClient) AddIssueWorklog(ctx context.Context, issueID string, req *youtrack.CreateWorklogRequest) (*youtrack.WorkItem, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.AddIssueWorklog(ytCtx, issueID, req)
}

// GetUserWorklogs returns worklogs for a user
func (c *YouTrackClient) GetUserWorklogs(ctx context.Context, userID string, projectID string, startDate, endDate string, skip, top int) ([]*youtrack.WorkItem, error) {
	ytCtx := c.WithContext(ctx)
	return c.client.GetUserWorklogs(ytCtx, userID, projectID, startDate, endDate, skip, top)
}

// GetKeyHash returns the hash of the API key for logging purposes
func (c *YouTrackClient) GetKeyHash(ctx context.Context) string {
	apiKey := c.GetEffectiveAPIKey(ctx)
	return logging.HashAPIKey(apiKey)
}

// GetAppLogger returns the app logger
func (c *YouTrackClient) GetAppLogger() *logging.AppLogger {
	return c.appLogger
}
