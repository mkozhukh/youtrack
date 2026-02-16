package handlers

import (
	"context"
	"fmt"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// UserHandlers manages user-related MCP operations
type UserHandlers struct {
	ytClient       UserClient
	defaultProject string
	toolLogger     func(string, map[string]interface{})
	errorHandler   *ErrorHandler
	projectTracker ProjectTracker
}

// UserClient defines the interface for YouTrack client operations needed for user management
type UserClient interface {
	GetCurrentUser(ctx context.Context) (*youtrack.User, error)
	GetProjectUsers(ctx context.Context, projectID string, skip, top int) ([]*youtrack.User, error)
}

// ProjectTracker defines the interface for tracking project usage
type ProjectTracker interface {
	TrackProject(ctx context.Context, projectID string)
	GetLastProject(ctx context.Context) string
}

// NewUserHandlers creates a new instance of UserHandlers
func NewUserHandlers(ytClient UserClient, defaultProject string, toolLogger func(string, map[string]interface{}), projectTracker ProjectTracker) *UserHandlers {
	return &UserHandlers{
		ytClient:       ytClient,
		defaultProject: defaultProject,
		toolLogger:     toolLogger,
		errorHandler:   NewErrorHandler(),
		projectTracker: projectTracker,
	}
}

// GetCurrentUserHandler handles the get_current_user tool call
func (h *UserHandlers) GetCurrentUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if h.toolLogger != nil {
		h.toolLogger("get_current_user", map[string]interface{}{})
	}

	user, err := h.ytClient.GetCurrentUser(ctx)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving current user"), nil
	}

	response := fmt.Sprintf("User Profile:\n")
	response += fmt.Sprintf("- ID: %s\n", user.ID)
	response += fmt.Sprintf("- Login: %s\n", user.Login)
	response += fmt.Sprintf("- Full Name: %s\n", user.FullName)
	response += fmt.Sprintf("- Email: %s\n", user.Email)

	if h.defaultProject != "" {
		response += fmt.Sprintf("- Default Project: %s\n", h.defaultProject)
	}

	if h.projectTracker != nil {
		if lastProject := h.projectTracker.GetLastProject(ctx); lastProject != "" {
			response += fmt.Sprintf("- Last Used Project: %s\n", lastProject)
		}
	}

	return mcp.NewToolResultText(response), nil
}

// GetProjectUsersHandler handles the get_project_users tool call
func (h *UserHandlers) GetProjectUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, err := request.RequireString("project_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("project_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("get_project_users", map[string]interface{}{
			"project_id": projectID,
		})
	}

	// Track project usage
	if h.projectTracker != nil {
		h.projectTracker.TrackProject(ctx, projectID)
	}

	// Fetch all users with pagination
	var allUsers []*youtrack.User
	skip := 0
	top := 100

	for {
		users, err := h.ytClient.GetProjectUsers(ctx, projectID, skip, top)
		if err != nil {
			return h.errorHandler.HandleError(err, "retrieving project users"), nil
		}

		if len(users) == 0 {
			break
		}

		allUsers = append(allUsers, users...)

		if len(users) < top {
			break
		}

		skip += len(users)
	}

	if len(allUsers) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No users found in project '%s'.", projectID)), nil
	}

	response := fmt.Sprintf("Project Members (%d):\n\n", len(allUsers))
	for _, user := range allUsers {
		response += fmt.Sprintf("- %s (%s)\n", user.FullName, user.Login)
		if user.Email != "" {
			response += fmt.Sprintf("  Email: %s\n", user.Email)
		}
	}

	return mcp.NewToolResultText(response), nil
}
