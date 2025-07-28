package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// CommentHandlers manages comment-related MCP operations
type CommentHandlers struct {
	ytClient     CommentClient
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
}

// CommentClient defines the interface for YouTrack client operations needed for comment management
type CommentClient interface {
	AddIssueComment(ctx context.Context, issueID string, comment string) (*youtrack.IssueComment, error)
	GetIssue(ctx context.Context, issueID string) (*youtrack.Issue, error)
}

// NewCommentHandlers creates a new instance of CommentHandlers
func NewCommentHandlers(ytClient CommentClient, toolLogger func(string, map[string]interface{})) *CommentHandlers {
	return &CommentHandlers{
		ytClient:     ytClient,
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
	}
}

// AddCommentHandler handles adding comments to issues
func (h *CommentHandlers) AddCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	commentText, err := request.RequireString("comment")
	if err != nil {
		return h.errorHandler.FormatValidationError("comment", err), nil
	}

	// Additional validation for comment text
	if err := h.errorHandler.ValidateRequiredParameter(commentText, "comment"); err != nil {
		return h.errorHandler.FormatValidationError("comment", err), nil
	}

	// Log the tool call
	if h.toolLogger != nil {
		h.toolLogger("add_comment", map[string]interface{}{
			"issue_id": issueID,
			"comment":  commentText,
		})
	}

	// Verify the issue exists first
	issue, err := h.ytClient.GetIssue(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "finding issue"), nil
	}

	// Add the comment to the issue
	comment, err := h.ytClient.AddIssueComment(ctx, issueID, commentText)
	if err != nil {
		return h.errorHandler.HandleError(err, "adding comment to issue"), nil
	}

	// Prepare the response
	details := fmt.Sprintf("Issue ID: %s\n", issueID)
	details += fmt.Sprintf("Issue: %s\n", issue.Summary)
	details += fmt.Sprintf("Comment ID: %s\n", comment.ID)
	details += fmt.Sprintf("Author: %s\n", comment.Author.Login)
	details += fmt.Sprintf("Created: %s\n", comment.Created.Format("2006-01-02 15:04:05"))
	details += fmt.Sprintf("\nComment text:\n%s", comment.Text)

	response := h.formatSuccessResult("Comment added successfully!", details)
	return mcp.NewToolResultText(response), nil
}

// formatSuccessResult formats a successful operation result
func (h *CommentHandlers) formatSuccessResult(title, details string) string {
	response := fmt.Sprintf("✅ %s\n", title)
	response += fmt.Sprintf("📋 %s\n", details)
	response += fmt.Sprintf("🕐 Completed at: %s\n", h.getCurrentTimestamp())
	return response
}

// getCurrentTimestamp returns the current timestamp in a readable format
func (h *CommentHandlers) getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}
