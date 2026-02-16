package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// LinkHandlers manages link-related MCP operations
type LinkHandlers struct {
	ytClient     LinkClient
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
}

// LinkClient defines the interface for YouTrack client operations needed for link management
type LinkClient interface {
	GetIssueLinks(ctx context.Context, issueID string) ([]*youtrack.IssueLink, error)
	CreateIssueLink(ctx context.Context, sourceID, targetID, linkType string) error
}

// NewLinkHandlers creates a new instance of LinkHandlers
func NewLinkHandlers(ytClient LinkClient, toolLogger func(string, map[string]interface{})) *LinkHandlers {
	return &LinkHandlers{
		ytClient:     ytClient,
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
	}
}

// GetIssueLinksHandler handles the get_issue_links tool call
func (h *LinkHandlers) GetIssueLinksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("get_issue_links", map[string]interface{}{
			"issue_id": issueID,
		})
	}

	links, err := h.ytClient.GetIssueLinks(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving issue links"), nil
	}

	if len(links) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No links found for issue %s.", issueID)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Links for %s:\n\n", issueID))

	for _, link := range links {
		typeName := "Unknown"
		if link.LinkType != nil {
			typeName = link.LinkType.Name
		}

		sb.WriteString(fmt.Sprintf("- %s (%s):\n", typeName, link.Direction))
		for _, issue := range link.Issues {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", issue.ID, issue.Summary))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// CreateIssueLinkHandler handles the create_issue_link tool call
func (h *LinkHandlers) CreateIssueLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourceID, err := request.RequireString("source_issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("source_issue_id", err), nil
	}

	targetID, err := request.RequireString("target_issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("target_issue_id", err), nil
	}

	linkType, err := request.RequireString("link_type")
	if err != nil {
		return h.errorHandler.FormatValidationError("link_type", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("create_issue_link", map[string]interface{}{
			"source_issue_id": sourceID,
			"target_issue_id": targetID,
			"link_type":       linkType,
		})
	}

	err = h.ytClient.CreateIssueLink(ctx, sourceID, targetID, linkType)
	if err != nil {
		return h.errorHandler.HandleError(err, "creating issue link"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Link created: %s -[%s]-> %s", sourceID, linkType, targetID)), nil
}
