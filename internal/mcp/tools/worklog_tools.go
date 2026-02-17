package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// AddWorklogTool returns the MCP tool definition for adding worklogs to issues
func AddWorklogTool() mcp.Tool {
	return mcp.NewTool("add_worklog",
		mcp.WithDescription("Log work time on an issue"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to log work on"),
		),
		mcp.WithString("duration",
			mcp.Required(),
			mcp.Description("Duration of work (e.g., '30m', '2h', '1h 30m', '2d', '1w'). Units: w=weeks, d=days (8h), h=hours, m=minutes. Plain number treated as minutes."),
		),
		mcp.WithString("text",
			mcp.Description("Description of the work performed (optional)"),
		),
		mcp.WithString("date",
			mcp.Description("Date of the work in YYYY-MM-DD format (optional, defaults to today)"),
		),
		mcp.WithString("work_type",
			mcp.Description("Type of work (e.g., 'Development', 'Testing', 'Documentation') (optional)"),
		),
	)
}

// GetIssueWorklogsTool returns the MCP tool definition for getting issue worklogs
func GetIssueWorklogsTool() mcp.Tool {
	return mcp.NewTool("get_issue_worklogs",
		mcp.WithDescription("Get all work items logged on a specific issue"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to retrieve worklogs for"),
		),
	)
}

// GetUserWorklogsTool returns the MCP tool definition for getting user worklogs
func GetUserWorklogsTool() mcp.Tool {
	return mcp.NewTool("get_user_worklogs",
		mcp.WithDescription("Get work items logged by a specific user"),
		mcp.WithString("user_id",
			mcp.Description("User ID to retrieve worklogs for (optional, defaults to current user)"),
		),
		mcp.WithString("project_id",
			mcp.Description("Filter by project ID (optional)"),
		),
		mcp.WithString("start_date",
			mcp.Description("Start date in YYYY-MM-DD format (optional)"),
		),
		mcp.WithString("end_date",
			mcp.Description("End date in YYYY-MM-DD format (optional)"),
		),
	)
}
