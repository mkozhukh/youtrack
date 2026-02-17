package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/mkozhukh/youtrack/internal/mcp/filestore"
	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/mark3labs/mcp-go/mcp"
)

// AttachmentHandlers manages attachment-related MCP operations
type AttachmentHandlers struct {
	ytClient     AttachmentClient
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
	fileStore    *filestore.Store
	fileBaseURL  string
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

// NewAttachmentHandlersWithFileStore creates AttachmentHandlers with file server support
func NewAttachmentHandlersWithFileStore(ytClient AttachmentClient, toolLogger func(string, map[string]interface{}), store *filestore.Store, baseURL string) *AttachmentHandlers {
	return &AttachmentHandlers{
		ytClient:     ytClient,
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
		fileStore:    store,
		fileBaseURL:  baseURL,
	}
}

// FileServerEnabled returns true if the file server sidecar is active
func (h *AttachmentHandlers) FileServerEnabled() bool {
	return h.fileStore != nil
}

// GetFileBaseURL returns the public base URL for the file server.
func (h *AttachmentHandlers) GetFileBaseURL() string {
	return h.fileBaseURL
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

// GetIssueAttachmentContentHandler handles the get_issue_attachment_content tool call.
// When the file server is enabled, it downloads content from YouTrack and stores it in the
// file store, returning a local HTTP URL. When disabled, it returns the YouTrack attachment URL.
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

	if h.fileStore != nil {
		return h.getAttachmentContentViaFileStore(ctx, issueID, attachmentID)
	}

	return h.getAttachmentContentURL(ctx, issueID, attachmentID)
}

// getAttachmentContentViaFileStore downloads content from YT, stores in filestore, returns local URL
func (h *AttachmentHandlers) getAttachmentContentViaFileStore(ctx context.Context, issueID, attachmentID string) (*mcp.CallToolResult, error) {
	// Get attachment metadata first to know the filename
	attachments, err := h.ytClient.GetIssueAttachments(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving attachment metadata"), nil
	}

	var filename string
	for _, att := range attachments {
		if att.ID == attachmentID {
			filename = att.Name
			break
		}
	}
	if filename == "" {
		filename = attachmentID
	}

	// Download the content
	data, err := h.ytClient.GetIssueAttachmentContent(ctx, issueID, attachmentID)
	if err != nil {
		return h.errorHandler.HandleError(err, "downloading attachment content"), nil
	}

	// Store in file store
	fileID, err := h.fileStore.Put(data, filename)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to store file: %v", err)), nil
	}

	url := fmt.Sprintf("%s/mcpfiles/%s", h.fileBaseURL, fileID)
	response := fmt.Sprintf("Attachment available for download:\n\n")
	response += fmt.Sprintf("- Name: %s\n", filename)
	response += fmt.Sprintf("- Size: %d bytes\n", len(data))
	response += fmt.Sprintf("- URL: %s\n", url)

	return mcp.NewToolResultText(response), nil
}

// getAttachmentContentURL returns the YouTrack native URL for the attachment
func (h *AttachmentHandlers) getAttachmentContentURL(ctx context.Context, issueID, attachmentID string) (*mcp.CallToolResult, error) {
	attachments, err := h.ytClient.GetIssueAttachments(ctx, issueID)
	if err != nil {
		return h.errorHandler.HandleError(err, "retrieving attachment metadata"), nil
	}

	for _, att := range attachments {
		if att.ID == attachmentID {
			response := fmt.Sprintf("Attachment download URL:\n\n")
			response += fmt.Sprintf("- Name: %s\n", att.Name)
			response += fmt.Sprintf("- Size: %d bytes\n", att.Size)
			if att.MimeType != "" {
				response += fmt.Sprintf("- Type: %s\n", att.MimeType)
			}
			response += fmt.Sprintf("- URL: %s\n", att.URL)
			return mcp.NewToolResultText(response), nil
		}
	}

	return mcp.NewToolResultError(fmt.Sprintf("Attachment %s not found on issue %s", attachmentID, issueID)), nil
}

// UploadAttachmentHandler handles the upload_attachment tool call.
// When the file server is enabled, it reads the file from the store by file_id.
// When disabled, it accepts base64-encoded content directly.
func (h *AttachmentHandlers) UploadAttachmentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueID, err := request.RequireString("issue_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("issue_id", err), nil
	}

	filename, err := request.RequireString("filename")
	if err != nil {
		return h.errorHandler.FormatValidationError("filename", err), nil
	}

	if h.fileStore != nil {
		return h.uploadAttachmentFromFileStore(ctx, request, issueID, filename)
	}

	return h.uploadAttachmentFromBase64(ctx, request, issueID, filename)
}

// uploadAttachmentFromFileStore reads file from store and uploads to YT
func (h *AttachmentHandlers) uploadAttachmentFromFileStore(ctx context.Context, request mcp.CallToolRequest, issueID, filename string) (*mcp.CallToolResult, error) {
	fileID, err := request.RequireString("file_id")
	if err != nil {
		return h.errorHandler.FormatValidationError("file_id", err), nil
	}

	if h.toolLogger != nil {
		h.toolLogger("upload_attachment", map[string]interface{}{
			"issue_id": issueID,
			"filename": filename,
			"file_id":  fileID,
		})
	}

	filePath, _, _, err := h.fileStore.Get(fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve file from store: %v", err)), nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read stored file: %v", err)), nil
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

// uploadAttachmentFromBase64 decodes base64 content and uploads to YT (legacy mode)
func (h *AttachmentHandlers) uploadAttachmentFromBase64(ctx context.Context, request mcp.CallToolRequest, issueID, filename string) (*mcp.CallToolResult, error) {
	contentB64, err := request.RequireString("content")
	if err != nil {
		return h.errorHandler.FormatValidationError("content", err), nil
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
