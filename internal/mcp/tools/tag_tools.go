package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// SearchTagsTool returns the MCP tool definition for searching tags
func SearchTagsTool() mcp.Tool {
	return mcp.NewTool("search_tags",
		mcp.WithDescription("Search for tags by partial name match (case-insensitive)"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Partial tag name to search for"),
		),
	)
}

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

// UntagIssueTool returns the MCP tool definition for removing tags from issues
func UntagIssueTool() mcp.Tool {
	return mcp.NewTool("untag_issue",
		mcp.WithDescription("Remove a tag from an issue in YouTrack"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to remove the tag from"),
		),
		mcp.WithString("tag",
			mcp.Required(),
			mcp.Description("Tag name to remove from the issue"),
		),
	)
}
