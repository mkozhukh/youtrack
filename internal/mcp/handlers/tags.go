package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// TagHandlers manages tag-related MCP operations
type TagHandlers struct {
	ytClient     YTClient
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
}

// YTClient defines the interface for YouTrack client operations needed for tag management
type YTClient interface {
	EnsureTag(ctx context.Context, tagName string, color string) (string, error)
	AddIssueTag(ctx context.Context, issueID string, tagName string) error
	GetIssue(ctx context.Context, issueID string) (*youtrack.Issue, error)
}

// NewTagHandlers creates a new instance of TagHandlers
func NewTagHandlers(ytClient YTClient, toolLogger func(string, map[string]interface{})) *TagHandlers {
	return &TagHandlers{
		ytClient:     ytClient,
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
	}
}

// TagIssueHandler handles adding tags to issues
func (h *TagHandlers) TagIssueHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	tagName, err := request.RequireString("tag")
	if err != nil {
		return h.errorHandler.FormatValidationError("tag", err), nil
	}

	// Additional validation for tag name
	if err := h.errorHandler.ValidateRequiredParameter(tagName, "tag"); err != nil {
		return h.errorHandler.FormatValidationError("tag", err), nil
	}

	// Log the tool call
	if h.toolLogger != nil {
		h.toolLogger("tag_issue", map[string]interface{}{
			"issue_id": issueID,
			"tag":      tagName,
		})
	}

	// Ensure the tag exists (create if needed)
	tagID, err := h.ytClient.EnsureTag(ctx, tagName, "")
	if err != nil {
		return h.errorHandler.HandleError(err, "ensuring tag exists"), nil
	}

	// Add the tag to the issue
	err = h.ytClient.AddIssueTag(ctx, issueID, tagName)
	if err != nil {
		return h.errorHandler.HandleError(err, "adding tag to issue"), nil
	}

	// Get updated issue details to confirm
	issue, err := h.ytClient.GetIssue(ctx, issueID)
	if err != nil {
		// Tag was added, but we couldn't get updated details
		return mcp.NewToolResultText(fmt.Sprintf("Tag '%s' (ID: %s) added to issue %s successfully", tagName, tagID, issueID)), nil
	}

	// Prepare response with updated tag list
	var tagNames []string
	for _, tag := range issue.Tags {
		tagNames = append(tagNames, tag.Name)
	}

	details := fmt.Sprintf("Issue ID: %s\n", issueID)
	details += fmt.Sprintf("Tag added: %s\n", tagName)
	details += fmt.Sprintf("Tag ID: %s\n", tagID)
	details += fmt.Sprintf("Current tags: %s\n", strings.Join(tagNames, ", "))

	response := h.formatSuccessResult("Tag added successfully!", details)
	return mcp.NewToolResultText(response), nil
}

// formatSuccessResult formats a successful operation result
func (h *TagHandlers) formatSuccessResult(title, details string) string {
	response := fmt.Sprintf("✅ %s\n", title)
	response += fmt.Sprintf("📋 %s\n", details)
	response += fmt.Sprintf("🕐 Completed at: %s\n", h.getCurrentTimestamp())
	return response
}

// getCurrentTimestamp returns the current timestamp in a readable format
func (h *TagHandlers) getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}
