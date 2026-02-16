package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// DropCacheTool returns the MCP tool definition for dropping cached project metadata
func DropCacheTool() mcp.Tool {
	return mcp.NewTool("drop_cache",
		mcp.WithDescription("Drop cached project metadata (custom fields, users). Use to force refresh of cached data."),
		mcp.WithString("project_id",
			mcp.Description("Project ID to drop cache for. If empty, drops cache for all projects."),
		),
	)
}
