package tools

import (
	"fmt"

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

// GetIssueAttachmentContentTool returns the MCP tool definition for downloading attachment content.
// When fileBaseURL is non-empty (fileserver mode), the description explains that the tool returns a
// local HTTP URL that can be fetched with curl. When empty, it returns the YouTrack native URL.
func GetIssueAttachmentContentTool(fileBaseURL string) mcp.Tool {
	if fileBaseURL != "" {
		desc := fmt.Sprintf(
			"Download the content of a specific issue attachment. " +
				"Returns a temporary local URL where the file can be downloaded. " +
				"Use curl or wget to fetch the returned URL, e.g.: curl -o output.file <returned_url>",
		)
		return mcp.NewTool("get_issue_attachment_content",
			mcp.WithDescription(desc),
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

	return mcp.NewTool("get_issue_attachment_content",
		mcp.WithDescription("Get the download URL for a specific issue attachment. Returns the YouTrack URL for the attachment."),
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

// UploadAttachmentTool returns the MCP tool definition for uploading attachments via file server.
// The description includes the concrete file server URL and the expected request/response format.
func UploadAttachmentTool(fileBaseURL string) mcp.Tool {
	desc := fmt.Sprintf(
		"Upload an attachment to an issue. "+
			"Workflow: first upload the file to the file server, then call this tool with the returned file_id.\n"+
			"Step 1 - Upload file: curl -F file=@/path/to/file %s/mcpfiles — returns JSON {\"file_id\": \"<id>\"}.\n"+
			"Step 2 - Call this tool with the file_id, issue_id, and desired filename.",
		fileBaseURL,
	)
	return mcp.NewTool("upload_attachment",
		mcp.WithDescription(desc),
		mcp.WithString("issue_id",
			mcp.Required(),
			mcp.Description("Issue ID to attach the file to"),
		),
		mcp.WithString("file_id",
			mcp.Required(),
			mcp.Description("File ID returned by the file server after uploading via POST "+fileBaseURL+"/mcpfiles"),
		),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Name of the file to create"),
		),
	)
}

// UploadAttachmentBase64Tool returns the legacy upload tool that accepts base64 content directly.
func UploadAttachmentBase64Tool() mcp.Tool {
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
