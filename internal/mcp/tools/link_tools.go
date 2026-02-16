package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// GetIssueLinksTool returns the MCP tool definition for getting issue links
func GetIssueLinksTool() mcp.Tool {
	return mcp.NewTool("get_issue_links",
		mcp.WithDescription("Get all links for a specific issue, grouped by link type and direction"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to retrieve links for"),
		),
	)
}

// CreateIssueLinkTool returns the MCP tool definition for creating an issue link
func CreateIssueLinkTool() mcp.Tool {
	return mcp.NewTool("create_issue_link",
		mcp.WithDescription("Create a link between two issues"),
		mcp.WithString("source_issue_id",
			mcp.Required(),
			mcp.Description("Source issue ID"),
		),
		mcp.WithString("target_issue_id",
			mcp.Required(),
			mcp.Description("Target issue ID"),
		),
		mcp.WithString("link_type",
			mcp.Required(),
			mcp.Description("Link type name (e.g. 'depends on', 'is required for', 'relates to', 'parent for', 'subtask of')"),
		),
	)
}
