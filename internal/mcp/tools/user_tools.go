package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// GetCurrentUserTool returns the MCP tool definition for getting the current user
func GetCurrentUserTool() mcp.Tool {
	return mcp.NewTool("get_current_user",
		mcp.WithDescription("Get the authenticated user's profile information including default project"),
	)
}

// GetProjectUsersTool returns the MCP tool definition for listing project members
func GetProjectUsersTool() mcp.Tool {
	return mcp.NewTool("get_project_users",
		mcp.WithDescription("List all users who are members of a specific project"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID (short name) to retrieve users for"),
		),
	)
}
