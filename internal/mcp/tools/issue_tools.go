package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// GetIssueListTool returns the MCP tool definition for listing issues
func GetIssueListTool() mcp.Tool {
	return mcp.NewTool("get_issue_list",
		mcp.WithDescription("Retrieve a list of issues from YouTrack with optional filtering and sorting"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID to search issues in"),
		),
		mcp.WithString("query",
			mcp.Description("YouTrack query string for filtering issues (optional)"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return (optional, defaults to config value)"),
		),
		mcp.WithString("sort_by",
			mcp.Description("Field to sort by, e.g. 'created', 'updated', 'priority' (optional)"),
		),
		mcp.WithString("sort_order",
			mcp.Description("Sort order: 'asc' or 'desc' (optional, defaults to 'desc')"),
		),
	)
}

// GetIssueDetailsTool returns the MCP tool definition for getting issue details
func GetIssueDetailsTool() mcp.Tool {
	return mcp.NewTool("get_issue_details",
		mcp.WithDescription("Get detailed information about a specific issue including comments"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to retrieve details for"),
		),
	)
}

// CreateIssueTool returns the MCP tool definition for creating issues
func CreateIssueTool() mcp.Tool {
	return mcp.NewTool("create_issue",
		mcp.WithDescription("Create a new issue in YouTrack"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID where the issue should be created"),
		),
		mcp.WithString("summary",
			mcp.Required(),
			mcp.Description("Issue summary/title"),
		),
		mcp.WithString("description",
			mcp.Description("Issue description (optional)"),
		),
	)
}

// DeleteIssueTool returns the MCP tool definition for deleting issues
func DeleteIssueTool() mcp.Tool {
	return mcp.NewTool("delete_issue",
		mcp.WithDescription("Delete an issue from YouTrack"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to delete"),
		),
	)
}

// UpdateIssueTool returns the MCP tool definition for updating issues
func UpdateIssueTool() mcp.Tool {
	return mcp.NewTool("update_issue",
		mcp.WithDescription("Update an existing issue in YouTrack"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to update"),
		),
		mcp.WithString("state",
			mcp.Description("New state for the issue (optional)"),
		),
		mcp.WithString("assignee",
			mcp.Description("New assignee login/username for the issue (optional)"),
		),
		mcp.WithString("summary",
			mcp.Description("New summary for the issue (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New description for the issue (optional)"),
		),
	)
}
