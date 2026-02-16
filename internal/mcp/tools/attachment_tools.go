package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// GetIssueAttachmentsTool returns the MCP tool definition for listing issue attachments
func GetIssueAttachmentsTool() mcp.Tool {
	return mcp.NewTool("get_issue_attachments",
		mcp.WithDescription("List all attachments for a specific issue with metadata (name, size, mime type)"),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to retrieve attachments for"),
		),
	)
}

// GetIssueAttachmentContentTool returns the MCP tool definition for downloading attachment content
func GetIssueAttachmentContentTool() mcp.Tool {
	return mcp.NewTool("get_issue_attachment_content",
		mcp.WithDescription("Download the content of a specific issue attachment. Returns text for text files, base64 for binary files."),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID the attachment belongs to"),
		),
		mcp.WithString("attachment_id",
			mcp.Required(),
			mcp.Description("Attachment ID to download"),
		),
	)
}

// UploadAttachmentTool returns the MCP tool definition for uploading attachments
func UploadAttachmentTool() mcp.Tool {
	return mcp.NewTool("upload_attachment",
		mcp.WithDescription("Upload an attachment to an issue. Content must be base64-encoded. Max 10MB."),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to attach the file to"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Base64-encoded file content"),
		),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Name of the file to create"),
		),
	)
}
