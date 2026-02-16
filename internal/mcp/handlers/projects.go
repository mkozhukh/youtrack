package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// ProjectHandlers manages project-related MCP operations
type ProjectHandlers struct {
	ytClient       ProjectClient
	toolLogger     func(string, map[string]interface{})
	errorHandler   *ErrorHandler
	projectTracker ProjectTracker
}

// ProjectClient defines the interface for YouTrack client operations needed for project management
type ProjectClient interface {
	GetProject(ctx context.Context, projectID string) (*youtrack.Project, error)
	GetProjectByName(ctx context.Context, name string) (*youtrack.Project, error)
	ListProjects(ctx context.Context, skip, top int) ([]*youtrack.Project, error)
	GetProjectCustomFields(ctx context.Context, projectID string) ([]*youtrack.CustomField, error)
	GetCustomFieldAllowedValues(ctx context.Context, projectID string, fieldName string) ([]youtrack.AllowedValue, error)
	GetAvailableLinkTypes(ctx context.Context) ([]*youtrack.LinkType, error)
}

// NewProjectHandlers creates a new instance of ProjectHandlers
func NewProjectHandlers(ytClient ProjectClient, toolLogger func(string, map[string]interface{}), projectTracker ProjectTracker) *ProjectHandlers {
	return &ProjectHandlers{
		ytClient:       ytClient,
		toolLogger:     toolLogger,
		errorHandler:   NewErrorHandler(),
		projectTracker: projectTracker,
	}
}

// GetProjectInfoHandler handles the get_project_info tool call
func (h *ProjectHandlers) GetProjectInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, err := request.RequireString("project_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("project_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("get_project_info", map[string]interface{}{
			"project_id": projectID,
		})
	}

	// Track project usage
	if h.projectTracker != nil {
		h.projectTracker.TrackProject(ctx, projectID)
	}

	// Get project details
	project, err := h.ytClient.GetProject(ctx, projectID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving project"), nil
	}

	// Get custom fields
	fields, err := h.ytClient.GetProjectCustomFields(ctx, projectID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving project custom fields"), nil
	}

	// Build response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project: %s (%s)\n", project.Name, project.ShortName))
	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", project.Description))
	}
	sb.WriteString("\n## Custom Fields\n\n")

	for _, field := range fields {
		sb.WriteString(fmt.Sprintf("- %s (type: %s)\n", field.Name, field.Type))

		// Try to get allowed values for this field
		values, err := h.ytClient.GetCustomFieldAllowedValues(ctx, projectID, field.Name)
		if err == nil && len(values) > 0 {
			var valueNames []string
			for _, v := range values {
				valueNames = append(valueNames, v.Name)
			}
			sb.WriteString(fmt.Sprintf("  Allowed values: %s\n", strings.Join(valueNames, ", ")))
		}
	}

	// Get link types
	linkTypes, err := h.ytClient.GetAvailableLinkTypes(ctx)
	if err == nil && len(linkTypes) > 0 {
		sb.WriteString("\n## Link Types\n\n")
		for _, lt := range linkTypes {
			sb.WriteString(fmt.Sprintf("- %s\n", lt.Name))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// ListProjectsHandler handles the list_projects tool call
func (h *ProjectHandlers) ListProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	query, _ := args["query"].(string)

	if h.toolLogger != nil {
		h.toolLogger("list_projects", map[string]interface{}{
			"query": query,
		})
	}

	if query != "" {
		// Search by name
		project, err := h.ytClient.GetProjectByName(ctx, query)
		if err != nil {
			return h.errorHandler.HandleError(err, "searching for project"), nil
		}

		response := fmt.Sprintf("Found project:\n- %s (ID: %s, Short: %s)\n", project.Name, project.ID, project.ShortName)
		if project.Description != "" {
			response += fmt.Sprintf("  Description: %s\n", project.Description)
		}
		return mcp.NewToolResultText(response), nil
	}

	// List all projects
	projects, err := h.ytClient.ListProjects(ctx, 0, 50)
	if err != nil {
		return h.errorHandler.HandleError(err, "listing projects"), nil
	}

	if len(projects) == 0 {
		return mcp.NewToolResultText("No projects found."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Projects (%d):\n\n", len(projects)))
	for _, p := range projects {
		sb.WriteString(fmt.Sprintf("- %s (ID: %s, Short: %s)\n", p.Name, p.ID, p.ShortName))
		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("  Description: %s\n", p.Description))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}
