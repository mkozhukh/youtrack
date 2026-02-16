package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// AttachmentHandlers manages attachment-related MCP operations
type AttachmentHandlers struct {
	ytClient     AttachmentClient
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
}

// AttachmentClient defines the interface for YouTrack client operations needed for attachment management
type AttachmentClient interface {
	GetIssueAttachments(ctx context.Context, issueID string) ([]*youtrack.Attachment, error)
	GetIssueAttachmentContent(ctx context.Context, issueID string, attachmentID string) ([]byte, error)
	AddIssueAttachmentFromBytes(ctx context.Context, issueID string, content []byte, filename string) (*youtrack.Attachment, error)
}

// NewAttachmentHandlers creates a new instance of AttachmentHandlers
func NewAttachmentHandlers(ytClient AttachmentClient, toolLogger func(string, map[string]interface{})) *AttachmentHandlers {
	return &AttachmentHandlers{
		ytClient:     ytClient,
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
	}
}

// GetIssueAttachmentsHandler handles the get_issue_attachments tool call
func (h *AttachmentHandlers) GetIssueAttachmentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("get_issue_attachments", map[string]interface{}{
			"issue_id": issueID,
		})
	}

	attachments, err := h.ytClient.GetIssueAttachments(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving issue attachments"), nil
	}

	if len(attachments) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No attachments found for issue %s.", issueID)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Attachments for %s (%d):\n\n", issueID, len(attachments)))

	for _, att := range attachments {
		sb.WriteString(fmt.Sprintf("- %s\n", att.Name))
		sb.WriteString(fmt.Sprintf("  ID: %s\n", att.ID))
		sb.WriteString(fmt.Sprintf("  Size: %d bytes\n", att.Size))
		if att.MimeType != "" {
			sb.WriteString(fmt.Sprintf("  Type: %s\n", att.MimeType))
		}
		if att.Author != nil {
			sb.WriteString(fmt.Sprintf("  Author: %s\n", att.Author.Login))
		}
		sb.WriteString(fmt.Sprintf("  Created: %s\n", att.Created.Format("2006-01-02 15:04:05")))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// GetIssueAttachmentContentHandler handles the get_issue_attachment_content tool call
func (h *AttachmentHandlers) GetIssueAttachmentContentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	attachmentID, err := request.RequireString("attachment_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("attachment_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("get_issue_attachment_content", map[string]interface{}{
			"issue_id":      issueID,
			"attachment_id": attachmentID,
		})
	}

	data, err := h.ytClient.GetIssueAttachmentContent(ctx, issueID, attachmentID)
	if err != nil {
		return h.errorHandler.HandleError(err, "downloading attachment content"), nil
	}

	// If valid UTF-8 text, return as text; otherwise base64 encode
	if utf8.Valid(data) {
		return mcp.NewToolResultText(string(data)), nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return mcp.NewToolResultText(fmt.Sprintf("[base64-encoded binary content (%d bytes)]\n%s", len(data), encoded)), nil
}

// UploadAttachmentHandler handles the upload_attachment tool call
func (h *AttachmentHandlers) UploadAttachmentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	contentB64, err := request.RequireString("content")
	if err != nil {
		return h.errorHandler.FormatValidationError("content", err), nil
	}

	filename, err := request.RequireString("filename")
	if err != nil {
		return h.errorHandler.FormatValidationError("filename", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("upload_attachment", map[string]interface{}{
			"issue_id": issueID,
			"filename": filename,
			"size":     len(contentB64),
		})
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(contentB64)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid base64 content: %v", err)), nil
	}

	// Check size limit (10MB)
	maxSize := 10 * 1024 * 1024
	if len(content) > maxSize {
		return mcp.NewToolResultError(fmt.Sprintf("File too large: %d bytes (max %d bytes)", len(content), maxSize)), nil
	}

	attachment, err := h.ytClient.AddIssueAttachmentFromBytes(ctx, issueID, content, filename)
	if err != nil {
		return h.errorHandler.HandleError(err, "uploading attachment"), nil
	}

	response := fmt.Sprintf("Attachment uploaded successfully!\n\n")
	response += fmt.Sprintf("- Name: %s\n", attachment.Name)
	response += fmt.Sprintf("- ID: %s\n", attachment.ID)
	response += fmt.Sprintf("- Size: %d bytes\n", attachment.Size)
	if attachment.MimeType != "" {
		response += fmt.Sprintf("- Type: %s\n", attachment.MimeType)
	}

	return mcp.NewToolResultText(response), nil
}
