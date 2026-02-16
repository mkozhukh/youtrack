package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// WorklogHandlers manages worklog-related MCP operations
type WorklogHandlers struct {
	ytClient     WorklogClient
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
}

// WorklogClient defines the interface for YouTrack client operations needed for worklog management
type WorklogClient interface {
	GetIssueWorklogs(ctx context.Context, issueID string) ([]*youtrack.WorkItem, error)
	AddIssueWorklog(ctx context.Context, issueID string, req *youtrack.CreateWorklogRequest) (*youtrack.WorkItem, error)
	GetUserWorklogs(ctx context.Context, userID string, projectID string, startDate, endDate string, skip, top int) ([]*youtrack.WorkItem, error)
	GetCurrentUser(ctx context.Context) (*youtrack.User, error)
}

// NewWorklogHandlers creates a new instance of WorklogHandlers
func NewWorklogHandlers(ytClient WorklogClient, toolLogger func(string, map[string]interface{})) *WorklogHandlers {
	return &WorklogHandlers{
		ytClient:     ytClient,
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
	}
}

// AddWorklogHandler handles the add_worklog tool call
func (h *WorklogHandlers) AddWorklogHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	args := request.GetArguments()
	durationFloat, ok := args["duration"].(float64)
	if !ok || durationFloat <= 0 {
		return mcp.NewToolResultError("duration is required and must be a positive number (minutes)"), nil
	}
	duration := int(durationFloat)

	text, _ := args["text"].(string)
	dateStr, _ := args["date"].(string)
	workType, _ := args["work_type"].(string)

	if h.toolLogger != nil {
		h.toolLogger("add_worklog", map[string]interface{}{
			"issue_id":  issueID,
			"duration":  duration,
			"text":      text,
			"date":      dateStr,
			"work_type": workType,
		})
	}

	req := &youtrack.CreateWorklogRequest{
		Duration:    duration,
		Description: text,
	}

	// Parse date if provided
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid date format '%s'. Use YYYY-MM-DD.", dateStr)), nil
		}
		dateMs := parsedDate.UnixNano() / int64(time.Millisecond)
		req.Date = &dateMs
	}

	// Set work type if provided
	if workType != "" {
		req.Type = &youtrack.WorkTypeRequest{Name: workType}
	}

	workItem, err := h.ytClient.AddIssueWorklog(ctx, issueID, req)
	if err != nil {
		return h.errorHandler.HandleError(err, "adding worklog"), nil
	}

	response := fmt.Sprintf("Worklog added successfully!\n\n")
	response += fmt.Sprintf("- Issue: %s\n", issueID)
	response += fmt.Sprintf("- Duration: %s\n", formatDuration(workItem.Duration))
	response += fmt.Sprintf("- Date: %s\n", workItem.Date.Format("2006-01-02"))
	if workItem.Description != "" {
		response += fmt.Sprintf("- Description: %s\n", workItem.Description)
	}
	if workItem.Type != nil {
		response += fmt.Sprintf("- Type: %s\n", workItem.Type.Name)
	}
	if workItem.Author != nil {
		response += fmt.Sprintf("- Author: %s\n", workItem.Author.Login)
	}

	return mcp.NewToolResultText(response), nil
}

// GetIssueWorklogsHandler handles the get_issue_worklogs tool call
func (h *WorklogHandlers) GetIssueWorklogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("get_issue_worklogs", map[string]interface{}{
			"issue_id": issueID,
		})
	}

	worklogs, err := h.ytClient.GetIssueWorklogs(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving issue worklogs"), nil
	}

	if len(worklogs) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No worklogs found for issue %s.", issueID)), nil
	}

	return mcp.NewToolResultText(h.formatWorklogs(worklogs, issueID)), nil
}

// GetUserWorklogsHandler handles the get_user_worklogs tool call
func (h *WorklogHandlers) GetUserWorklogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID, _ := args["user_id"].(string)
	projectID, _ := args["project_id"].(string)
	startDate, _ := args["start_date"].(string)
	endDate, _ := args["end_date"].(string)

	// If no user_id provided, get current user
	if userID == "" {
		currentUser, err := h.ytClient.GetCurrentUser(ctx)
		if err != nil {
			return h.errorHandler.HandleError(err, "getting current user"), nil
		}
		userID = currentUser.ID
	}

	if h.toolLogger != nil {
		h.toolLogger("get_user_worklogs", map[string]interface{}{
			"user_id":    userID,
			"project_id": projectID,
			"start_date": startDate,
			"end_date":   endDate,
		})
	}

	// Fetch all worklogs with pagination
	var allWorklogs []*youtrack.WorkItem
	skip := 0
	top := 100

	for {
		worklogs, err := h.ytClient.GetUserWorklogs(ctx, userID, projectID, startDate, endDate, skip, top)
		if err != nil {
			return h.errorHandler.HandleError(err, "retrieving user worklogs"), nil
		}

		if len(worklogs) == 0 {
			break
		}

		allWorklogs = append(allWorklogs, worklogs...)

		if len(worklogs) < top {
			break
		}

		skip += len(worklogs)
	}

	if len(allWorklogs) == 0 {
		return mcp.NewToolResultText("No worklogs found for the specified criteria."), nil
	}

	return mcp.NewToolResultText(h.formatUserWorklogs(allWorklogs, userID)), nil
}

func (h *WorklogHandlers) formatWorklogs(worklogs []*youtrack.WorkItem, issueID string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Worklogs for %s (%d):\n\n", issueID, len(worklogs)))

	totalMinutes := 0
	for _, wl := range worklogs {
		totalMinutes += wl.Duration
		sb.WriteString(fmt.Sprintf("- %s | %s", wl.Date.Format("2006-01-02"), formatDuration(wl.Duration)))
		if wl.Author != nil {
			sb.WriteString(fmt.Sprintf(" | %s", wl.Author.Login))
		}
		if wl.Type != nil {
			sb.WriteString(fmt.Sprintf(" | %s", wl.Type.Name))
		}
		sb.WriteString("\n")
		if wl.Description != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", wl.Description))
		}
	}

	sb.WriteString(fmt.Sprintf("\nTotal time logged: %s\n", formatDuration(totalMinutes)))

	return sb.String()
}

func (h *WorklogHandlers) formatUserWorklogs(worklogs []*youtrack.WorkItem, userID string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Worklogs for user %s (%d):\n\n", userID, len(worklogs)))

	totalMinutes := 0
	for _, wl := range worklogs {
		totalMinutes += wl.Duration
		sb.WriteString(fmt.Sprintf("- %s | %s", wl.Date.Format("2006-01-02"), formatDuration(wl.Duration)))
		if wl.Issue != nil {
			sb.WriteString(fmt.Sprintf(" | %s", wl.Issue.ID))
		}
		if wl.Type != nil {
			sb.WriteString(fmt.Sprintf(" | %s", wl.Type.Name))
		}
		sb.WriteString("\n")
		if wl.Description != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", wl.Description))
		}
	}

	sb.WriteString(fmt.Sprintf("\nTotal time logged: %s\n", formatDuration(totalMinutes)))

	return sb.String()
}

// formatDuration formats minutes into a human-readable duration string
func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}
