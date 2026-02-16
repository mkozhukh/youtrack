package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// GetProjectInfoTool returns the MCP tool definition for getting project schema info
func GetProjectInfoTool() mcp.Tool {
	return mcp.NewTool("get_project_info",
		mcp.WithDescription("Get project schema information including custom fields with allowed values and available link types"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID (short name) to retrieve info for"),
		),
	)
}

// ListProjectsTool returns the MCP tool definition for listing projects
func ListProjectsTool() mcp.Tool {
	return mcp.NewTool("list_projects",
		mcp.WithDescription("List available YouTrack projects, optionally searching by name"),
		mcp.WithString("query",
			mcp.Description("Optional project name to search for (case-insensitive)"),
		),
	)
}
