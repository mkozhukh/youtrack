package youtrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// GetIssueAttachments retrieves all attachments for a specific issue
func (c *Client) GetIssueAttachments(ctx *YouTrackContext, issueID string) ([]*Attachment, error) {
	path := fmt.Sprintf("/api/issues/%s/attachments", issueID)

	resp, err := c.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue attachments: %w", err)
	}
	defer resp.Body.Close()

	var attachments []*Attachment
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("failed to decode attachments: %w", err)
	}

	return attachments, nil
}

// AddIssueAttachment uploads a file as an attachment to an issue
func (c *Client) AddIssueAttachment(ctx *YouTrackContext, issueID string, filePath string) (*Attachment, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file field
	fileName := filepath.Base(filePath)
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content to form
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close the writer to finalize the form
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close form writer: %w", err)
	}

	// Make the API request
	path := fmt.Sprintf("/api/issues/%s/attachments", issueID)
	resp, err := c.doMultipartRequest(ctx, http.MethodPost, path, &requestBody, writer.FormDataContentType())
	if err != nil {
		return nil, fmt.Errorf("failed to upload attachment: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var attachments []*Attachment
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("failed to decode attachment response: %w", err)
	}

	// Find the newly created attachment (it should be the one with matching name and size)
	for _, attachment := range attachments {
		if attachment.Name == fileName && attachment.Size == fileInfo.Size() {
			return attachment, nil
		}
	}

	// If we can't find a matching attachment, return the first one (fallback)
	if len(attachments) > 0 {
		return attachments[0], nil
	}

	return nil, fmt.Errorf("uploaded attachment not found in response")
}

// doMultipartRequest makes an HTTP request with multipart form data
func (c *Client) doMultipartRequest(ctx *YouTrackContext, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	fullURL := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx.Context(), method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ctx.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	return resp, nil
}
