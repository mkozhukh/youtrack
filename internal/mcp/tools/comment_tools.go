package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// AddCommentTool returns the MCP tool definition for adding comments to issues
func AddCommentTool() mcp.Tool {
	return mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to an issue in YouTrack"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to add the comment to"),
		),
		mcp.WithString("comment",
			mcp.Required(),
			mcp.Description("Comment text to add to the issue"),
		),
	)
}
