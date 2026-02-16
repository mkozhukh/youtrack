package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mkozhukh/youtrack/internal/mcp/resolver"
	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
)

// IssueHandlers contains all the handlers for issue-related tools
type IssueHandlers struct {
	ytClient       YouTrackClientInterface
	resolver       *resolver.Resolver
	toolLogger     func(string, map[string]interface{})
	errorHandler   *ErrorHandler
	projectTracker ProjectTracker
}

// YouTrackClientInterface defines the interface for YouTrack client operations
type YouTrackClientInterface interface {
	SearchIssues(ctx context.Context, query string, skip, top int) ([]*youtrack.Issue, error)
	SearchIssuesSorted(ctx context.Context, query string, skip, top int, sortBy, sortOrder string) ([]*youtrack.Issue, error)
	GetIssue(ctx context.Context, issueID string) (*youtrack.Issue, error)
	GetIssueComments(ctx context.Context, issueID string) ([]*youtrack.IssueComment, error)
	GetIssueCustomFields(ctx context.Context, issueID string) ([]*youtrack.CustomFieldValue, error)
	CreateIssue(ctx context.Context, req *youtrack.CreateIssueRequest) (*youtrack.Issue, error)
	UpdateIssue(ctx context.Context, issueID string, req *youtrack.UpdateIssueRequest) (*youtrack.Issue, error)
	UpdateIssueAssigneeByProject(ctx context.Context, issueID string, projectID string, username string) (*youtrack.Issue, error)
	DeleteIssue(ctx context.Context, issueID string) error
	// Resolver support
	GetProjectUsers(ctx context.Context, projectID string, skip, top int) ([]*youtrack.User, error)
	GetCustomFieldAllowedValues(ctx context.Context, projectID string, fieldName string) ([]youtrack.AllowedValue, error)
}

// NewIssueHandlers creates a new instance of IssueHandlers
func NewIssueHandlers(ytClient YouTrackClientInterface, toolLogger func(string, map[string]interface{}), projectTracker ProjectTracker) *IssueHandlers {
	return &IssueHandlers{
		ytClient:       ytClient,
		resolver:       resolver.NewResolver(ytClient),
		toolLogger:     toolLogger,
		errorHandler:   NewErrorHandler(),
		projectTracker: projectTracker,
	}
}

// GetIssueListHandler handles the get_issue_list tool call
func (h *IssueHandlers) GetIssueListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	projectID, err := request.RequireString("project_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("project_id", err), nil
	}

	args := request.GetArguments()
	query, _ := args["query"].(string)
	maxResults, _ := args["max_results"].(float64)
	sortBy, _ := args["sort_by"].(string)
	sortOrder, _ := args["sort_order"].(string)

	// Convert max_results to int and validate
	maxResultsInt := int(maxResults)
	if maxResults > 0 {
		if err := h.errorHandler.ValidatePositiveNumber(maxResults, "max_results"); err != nil {
			return h.errorHandler.FormatValidationError("max_results", err), nil
		}
	}

	// Default sort order
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Track project usage
	if h.projectTracker != nil {
		h.projectTracker.TrackProject(ctx, projectID)
	}

	// Build optimized query with smart defaults
	optimizedQuery := h.buildOptimizedQuery(projectID, query)

	// Log the tool call
	if h.toolLogger != nil {
		h.toolLogger("get_issue_list", map[string]interface{}{
			"project_id":      projectID,
			"query":           query,
			"optimized_query": optimizedQuery,
			"max_results":     maxResultsInt,
			"sort_by":         sortBy,
			"sort_order":      sortOrder,
		})
	}

	// Search for issues — use sorted search if sort_by is provided
	var issues []*youtrack.Issue
	if sortBy != "" {
		issues, err = h.ytClient.SearchIssuesSorted(ctx, optimizedQuery, 0, maxResultsInt, sortBy, sortOrder)
	} else {
		issues, err = h.ytClient.SearchIssues(ctx, optimizedQuery, 0, maxResultsInt)
	}
	if err != nil {
		return h.errorHandler.HandleError(err, "searching issues"), nil
	}

	// Format the response
	response := h.formatIssueList(issues)
	return mcp.NewToolResultText(response), nil
}

// GetIssueDetailsHandler handles the get_issue_details tool call
func (h *IssueHandlers) GetIssueDetailsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	// Log the tool call
	if h.toolLogger != nil {
		h.toolLogger("get_issue_details", map[string]interface{}{
			"issue_id": issueID,
		})
	}

	// Get the issue details
	issue, err := h.ytClient.GetIssue(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving issue details"), nil
	}

	// Get the issue comments
	comments, err := h.ytClient.GetIssueComments(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving issue comments"), nil
	}

	// Get custom fields (non-critical, log error but continue)
	customFields, err := h.ytClient.GetIssueCustomFields(ctx, issueID)
	if err != nil {
		log.Warn("Failed to retrieve custom fields", "issue_id", issueID, "error", err)
	}

	// Format the response
	response := h.formatIssueDetails(issue, comments, customFields)
	return mcp.NewToolResultText(response), nil
}

// CreateIssueHandler handles the create_issue tool call
func (h *IssueHandlers) CreateIssueHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	projectID, err := request.RequireString("project_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("project_id", err), nil
	}

	summary, err := request.RequireString("summary")
	if err != nil {
		return h.errorHandler.FormatValidationError("summary", err), nil
	}

	// Additional validation for summary
	if err := h.errorHandler.ValidateRequiredParameter(summary, "summary"); err != nil {
		return h.errorHandler.FormatValidationError("summary", err), nil
	}

	args := request.GetArguments()
	description, _ := args["description"].(string)

	// Track project usage
	if h.projectTracker != nil {
		h.projectTracker.TrackProject(ctx, projectID)
	}

	// Log the tool call
	if h.toolLogger != nil {
		h.toolLogger("create_issue", map[string]interface{}{
			"project_id":  projectID,
			"summary":     summary,
			"description": description,
		})
	}

	// Create the issue request
	createReq := &youtrack.CreateIssueRequest{
		Project:     youtrack.ProjectRef{ID: projectID},
		Summary:     summary,
		Description: description,
	}

	// Create the issue
	issue, err := h.ytClient.CreateIssue(ctx, createReq)
	if err != nil {
		return h.errorHandler.HandleError(err, "creating issue"), nil
	}

	// Format the response
	response := h.formatCreatedIssue(issue)
	return mcp.NewToolResultText(response), nil
}

// UpdateIssueHandler handles the update_issue tool call
func (h *IssueHandlers) UpdateIssueHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	args := request.GetArguments()
	state, _ := args["state"].(string)
	assignee, _ := args["assignee"].(string)
	summary, _ := args["summary"].(string)
	description, _ := args["description"].(string)

	// Log the tool call
	if h.toolLogger != nil {
		h.toolLogger("update_issue", map[string]interface{}{
			"issue_id":    issueID,
			"state":       state,
			"assignee":    assignee,
			"summary":     summary,
			"description": description,
		})
	}

	// Extract project ID from the issue ID (assuming format like PROJECT-123)
	projectID := extractProjectFromIssueID(issueID)
	if projectID == "" {
		return mcp.NewToolResultError("Could not extract project ID from issue ID. Issue ID should be in format PROJECT-123"), nil
	}

	// Start with getting the current issue
	currentIssue, err := h.ytClient.GetIssue(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving current issue"), nil
	}

	// Create the update request
	updateReq := &youtrack.UpdateIssueRequest{}
	hasUpdates := false

	if summary != "" {
		updateReq.Summary = &summary
		hasUpdates = true
	}

	if description != "" {
		updateReq.Description = &description
		hasUpdates = true
	}

	// Resolve and validate state if provided
	if state != "" {
		resolvedState, err := h.resolver.ResolveEnumValue(ctx, projectID, "State", state)
		if err != nil {
			if resolveErr, ok := err.(*resolver.ResolveError); ok {
				return mcp.NewToolResultError(resolveErr.Error()), nil
			}
			return h.errorHandler.HandleError(err, "resolving state value"), nil
		}

		updateReq.Fields = append(updateReq.Fields, youtrack.CustomField{
			Name:  "State",
			Type:  "StateIssueCustomField",
			Value: youtrack.SingleValue{Value: resolvedState},
		})
		hasUpdates = true
	}

	var updatedIssue *youtrack.Issue

	// Update the issue if there are basic updates
	if hasUpdates {
		updatedIssue, err = h.ytClient.UpdateIssue(ctx, issueID, updateReq)
		if err != nil {
			return h.errorHandler.HandleError(err, "updating issue"), nil
		}
	}

	// Handle assignee update separately if provided
	if assignee != "" {
		// Resolve assignee using smart matching
		resolvedAssignee, err := h.resolver.ResolveUser(ctx, projectID, assignee)
		if err != nil {
			if resolveErr, ok := err.(*resolver.ResolveError); ok {
				return mcp.NewToolResultError(resolveErr.Error()), nil
			}
			return h.errorHandler.HandleError(err, "resolving assignee"), nil
		}

		updatedIssue, err = h.ytClient.UpdateIssueAssigneeByProject(ctx, issueID, projectID, resolvedAssignee)
		if err != nil {
			return h.errorHandler.HandleError(err, "updating issue assignee"), nil
		}
	}

	// If no updates were made, return the current issue
	if updatedIssue == nil {
		updatedIssue = currentIssue
	}

	// Format the response
	response := h.formatUpdatedIssue(updatedIssue)
	return mcp.NewToolResultText(response), nil
}

// Helper functions for query optimization and formatting

// buildOptimizedQuery creates an optimized query with smart defaults
func (h *IssueHandlers) buildOptimizedQuery(projectID, userQuery string) string {
	// Start with project filter
	query := fmt.Sprintf("project: %s", projectID)

	// If user provided a query, add it to the project filter
	if userQuery != "" {
		query = fmt.Sprintf("%s %s", query, userQuery)
	} else {
		// Apply smart defaults when no query is provided
		query = fmt.Sprintf("%s updated: {Last week}", query)  // Show issues updated in last 30 days
		query = fmt.Sprintf("%s sort by: updated desc", query) // Sort by most recently updated
	}

	// Add sorting if not already present
	if !strings.Contains(query, "sort by:") {
		query = fmt.Sprintf("%s sort by: updated desc", query)
	}

	return query
}

// formatEmptyResult formats an empty result with helpful information
func (h *IssueHandlers) formatEmptyResult(title, suggestion string) string {
	response := fmt.Sprintf("❌ %s\n", title)
	response += fmt.Sprintf("💡 %s\n", suggestion)
	response += fmt.Sprintf("🕐 Checked at: %s\n", h.getCurrentTimestamp())
	return response
}

// formatSuccessResult formats a successful operation result
func (h *IssueHandlers) formatSuccessResult(title, details string) string {
	response := fmt.Sprintf("✅ %s\n", title)
	response += fmt.Sprintf("📋 %s\n", details)
	response += fmt.Sprintf("🕐 Completed at: %s\n", h.getCurrentTimestamp())
	return response
}

// getCurrentTimestamp returns the current timestamp in a readable format
func (h *IssueHandlers) getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}

func (h *IssueHandlers) formatIssueList(issues []*youtrack.Issue) string {
	if len(issues) == 0 {
		return h.formatEmptyResult("No issues found", "Try adjusting your query or search criteria")
	}

	// Create header with metadata
	header := fmt.Sprintf("📋 Issues Found: %d\n", len(issues))
	header += fmt.Sprintf("🔍 Search completed at: %s\n", h.getCurrentTimestamp())
	header += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	response := header
	for i, issue := range issues {
		assignee := "Unassigned"
		if issue.Assignee != nil {
			assignee = issue.Assignee.Login
		}

		reporter := "Unknown"
		if issue.Reporter != nil {
			reporter = issue.Reporter.Login
		}

		response += fmt.Sprintf("%d. 🎫 %s\n", i+1, issue.ID)
		response += fmt.Sprintf("   📝 Summary: %s\n", issue.Summary)
		response += fmt.Sprintf("   👤 Assignee: %s\n", assignee)
		response += fmt.Sprintf("   📩 Reporter: %s\n", reporter)
		response += fmt.Sprintf("   📅 Created: %s\n", issue.Created.Format("2006-01-02 15:04:05"))
		response += fmt.Sprintf("   🔄 Updated: %s\n", issue.Updated.Format("2006-01-02 15:04:05"))

		if len(issue.Tags) > 0 {
			response += "   🏷️  Tags: "
			for j, tag := range issue.Tags {
				if j > 0 {
					response += ", "
				}
				response += tag.Name
			}
			response += "\n"
		}

		response += "\n"
	}

	// Add footer with metadata
	footer := fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	footer += fmt.Sprintf("📊 Total issues: %d | 🔍 Fetched at: %s\n", len(issues), h.getCurrentTimestamp())

	return response + footer
}

func (h *IssueHandlers) formatIssueDetails(issue *youtrack.Issue, comments []*youtrack.IssueComment, customFields []*youtrack.CustomFieldValue) string {
	assignee := "Unassigned"
	if issue.Assignee != nil {
		assignee = fmt.Sprintf("%s (%s)", issue.Assignee.FullName, issue.Assignee.Login)
	}

	reporter := "Unknown"
	if issue.Reporter != nil {
		reporter = fmt.Sprintf("%s (%s)", issue.Reporter.FullName, issue.Reporter.Login)
	}

	// Create header with metadata
	header := fmt.Sprintf("🎫 Issue Details: %s\n", issue.ID)
	header += fmt.Sprintf("🔍 Retrieved at: %s\n", h.getCurrentTimestamp())
	header += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	response := header
	response += fmt.Sprintf("📝 Summary: %s\n", issue.Summary)
	response += fmt.Sprintf("📄 Description: %s\n", issue.Description)
	response += fmt.Sprintf("👤 Assignee: %s\n", assignee)
	response += fmt.Sprintf("📩 Reporter: %s\n", reporter)
	response += fmt.Sprintf("📅 Created: %s\n", issue.Created.Format("2006-01-02 15:04:05"))
	response += fmt.Sprintf("🔄 Updated: %s\n", issue.Updated.Format("2006-01-02 15:04:05"))

	if issue.Resolved != nil {
		response += fmt.Sprintf("✅ Resolved: %s\n", issue.Resolved.Format("2006-01-02 15:04:05"))
	}

	if len(issue.Tags) > 0 {
		response += "🏷️  Tags: "
		for i, tag := range issue.Tags {
			if i > 0 {
				response += ", "
			}
			response += tag.Name
		}
		response += "\n"
	}

	if len(customFields) > 0 {
		response += "\n📌 Custom Fields:\n"
		for _, field := range customFields {
			response += fmt.Sprintf("   %s: %v\n", field.Name, formatFieldValue(field.Value))
		}
	}

	if len(comments) > 0 {
		response += fmt.Sprintf("\n💬 Comments (%d):\n", len(comments))
		response += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
		for i, comment := range comments {
			author := "Unknown"
			if comment.Author != nil {
				author = comment.Author.Login
			}
			response += fmt.Sprintf("%d. 👤 %s (%s)\n", i+1, author, comment.Created.Format("2006-01-02 15:04:05"))
			response += fmt.Sprintf("   📝 %s\n\n", comment.Text)
		}
	}

	// Add footer with metadata
	footer := fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	footer += fmt.Sprintf("📊 Comments: %d | 🔍 Retrieved at: %s\n", len(comments), h.getCurrentTimestamp())

	return response + footer
}

func (h *IssueHandlers) formatCreatedIssue(issue *youtrack.Issue) string {
	details := fmt.Sprintf("Issue ID: %s\n", issue.ID)
	details += fmt.Sprintf("Summary: %s\n", issue.Summary)
	details += fmt.Sprintf("Description: %s\n", issue.Description)
	details += fmt.Sprintf("Created: %s\n", issue.Created.Format("2006-01-02 15:04:05"))

	if issue.Reporter != nil {
		details += fmt.Sprintf("Reporter: %s\n", issue.Reporter.Login)
	}

	return h.formatSuccessResult("Issue created successfully!", details)
}

func (h *IssueHandlers) formatUpdatedIssue(issue *youtrack.Issue) string {
	assignee := "Unassigned"
	if issue.Assignee != nil {
		assignee = issue.Assignee.Login
	}

	details := fmt.Sprintf("Issue ID: %s\n", issue.ID)
	details += fmt.Sprintf("Summary: %s\n", issue.Summary)
	details += fmt.Sprintf("Description: %s\n", issue.Description)
	details += fmt.Sprintf("Assignee: %s\n", assignee)
	details += fmt.Sprintf("Updated: %s\n", issue.Updated.Format("2006-01-02 15:04:05"))

	return h.formatSuccessResult("Issue updated successfully!", details)
}

// DeleteIssueHandler handles the delete_issue tool call
func (h *IssueHandlers) DeleteIssueHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("delete_issue", map[string]interface{}{
			"issue_id": issueID,
		})
	}

	err = h.ytClient.DeleteIssue(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "deleting issue"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Issue %s deleted successfully.", issueID)), nil
}

// formatFieldValue converts a custom field value to a display string
func formatFieldValue(value interface{}) string {
	if value == nil {
		return "<empty>"
	}
	switch v := value.(type) {
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			return name
		}
		if login, ok := v["login"].(string); ok {
			return login
		}
		return fmt.Sprintf("%v", v)
	case []interface{}:
		var names []string
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if name, ok := m["name"].(string); ok {
					names = append(names, name)
					continue
				}
			}
			names = append(names, fmt.Sprintf("%v", item))
		}
		return strings.Join(names, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// extractProjectFromIssueID extracts the project ID from an issue ID
// Assumes format like "PROJECT-123" -> "PROJECT"
func extractProjectFromIssueID(issueID string) string {
	for i, char := range issueID {
		if char == '-' {
			return issueID[:i]
		}
	}
	return ""
}
