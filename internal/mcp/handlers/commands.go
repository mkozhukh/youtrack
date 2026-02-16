package handlers

import (
	"context"
	"fmt"

	"github.com/mkozhukh/youtrack/internal/mcp/resolver"
	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// CommandHandlers manages command-related MCP operations
type CommandHandlers struct {
	ytClient     CommandClient
	resolver     *resolver.Resolver
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
}

// CommandClient defines the interface for YouTrack client operations needed for command execution
type CommandClient interface {
	ApplyCommand(ctx context.Context, issueID string, command string) error
	// Resolver support
	GetProjectUsers(ctx context.Context, projectID string, skip, top int) ([]*youtrack.User, error)
	GetCustomFieldAllowedValues(ctx context.Context, projectID string, fieldName string) ([]youtrack.AllowedValue, error)
}

// NewCommandHandlers creates a new instance of CommandHandlers
func NewCommandHandlers(ytClient CommandClient, toolLogger func(string, map[string]interface{})) *CommandHandlers {
	return &CommandHandlers{
		ytClient:     ytClient,
		resolver:     resolver.NewResolver(ytClient),
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
	}
}

// ApplyCommandHandler handles the apply_command tool call
func (h *CommandHandlers) ApplyCommandHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	command, err := request.RequireString("command")
	if err != nil {
		return h.errorHandler.FormatValidationError("command", err), nil
	}

	// Extract project ID from issue ID
	projectID := extractProjectFromIssueID(issueID)
	if projectID == "" {
		return mcp.NewToolResultError("Could not extract project ID from issue ID. Issue ID should be in format PROJECT-123"), nil
	}

	// Try to resolve field values in the command using smart matching
	resolvedCommand := command
	if resolver.IsResolvableCommand(command) {
		resolved, err := h.resolver.ResolveCommand(ctx, projectID, command)
		if err != nil {
			if resolveErr, ok := err.(*resolver.ResolveError); ok {
				return mcp.NewToolResultError(resolveErr.Error()), nil
			}
			return h.errorHandler.HandleError(err, "resolving command values"), nil
		}
		resolvedCommand = resolved
	}

	if h.toolLogger != nil {
		h.toolLogger("apply_command", map[string]interface{}{
			"issue_id":         issueID,
			"command":          command,
			"resolved_command": resolvedCommand,
		})
	}

	err = h.ytClient.ApplyCommand(ctx, issueID, resolvedCommand)
	if err != nil {
		return h.errorHandler.HandleError(err, "applying command"), nil
	}

	// Show resolved command if it was modified
	if resolvedCommand != command {
		return mcp.NewToolResultText(fmt.Sprintf("Command applied to issue %s successfully.\nOriginal: '%s'\nResolved: '%s'", issueID, command, resolvedCommand)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Command '%s' applied to issue %s successfully.", command, issueID)), nil
}
