package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// ApplyCommandTool returns the MCP tool definition for applying a YouTrack command to an issue
func ApplyCommandTool() mcp.Tool {
	return mcp.NewTool("apply_command",
		mcp.WithDescription("Execute a YouTrack command on an issue (e.g. 'State Open', 'Priority Critical', 'Type Bug')"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to apply the command to"),
		),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("YouTrack command string to execute"),
		),
	)
}
