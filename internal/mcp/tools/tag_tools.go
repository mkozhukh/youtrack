package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// TagIssueTool returns the MCP tool definition for tagging issues
func TagIssueTool() mcp.Tool {
	return mcp.NewTool("tag_issue",
		mcp.WithDescription("Add a tag to an issue in YouTrack. Creates the tag if it doesn't exist."),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to add the tag to"),
		),
		mcp.WithString("tag",
			mcp.Required(),
			mcp.Description("Tag name to add to the issue"),
		),
	)
}
