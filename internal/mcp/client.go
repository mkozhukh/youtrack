package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/charmbracelet/log"
)

// YouTrackClient wraps the YouTrack client with configuration and context handling
type YouTrackClient struct {
	client     *youtrack.Client
	config     YouTrackConfig
	defaultCtx *youtrack.YouTrackContext
}

// NewYouTrackClient creates a new YouTrack client wrapper with configuration
func NewYouTrackClient(config YouTrackConfig) (*YouTrackClient, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create the base client
	client := youtrack.NewClient(config.BaseURL)

	// Note: The client timeout is set in the NewClient constructor
	// We would need to modify the client constructor to support custom timeouts

	// Create default context
	defaultCtx := youtrack.NewYouTrackContext(context.Background(), config.APIKey)

	ytClient := &YouTrackClient{
		client:     client,
		config:     config,
		defaultCtx: defaultCtx,
	}

	// Test connection
	if err := ytClient.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to YouTrack: %w", err)
	}

	log.Info("YouTrack client initialized successfully", "base_url", config.BaseURL)
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

	if config.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	// Validate API key format (YouTrack API keys are typically long hex strings)
	if len(config.APIKey) < 10 {
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
func (c *YouTrackClient) WithContext(ctx context.Context) *youtrack.YouTrackContext {
	return youtrack.NewYouTrackContext(ctx, c.config.APIKey)
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

// AddIssueTag adds a tag to an issue
func (c *YouTrackClient) AddIssueTag(ctx context.Context, issueID string, tagName string) error {
	ytCtx := c.WithContext(ctx)
	return c.client.AddIssueTag(ytCtx, issueID, tagName)
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
