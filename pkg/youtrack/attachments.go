package youtrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// GetIssueAttachments retrieves all attachments for a specific issue
func (c *Client) GetIssueAttachments(ctx *YouTrackContext, issueID string) ([]*Attachment, error) {
	path := fmt.Sprintf("/api/issues/%s/attachments", issueID)

	query := url.Values{}
	query.Add("fields", "id,name,size,created,mimeType,url,author(id,login,fullName)")

	resp, err := c.Get(ctx, path, query)
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

// GetIssueAttachmentContent downloads the raw content of an attachment.
// It first fetches the attachment metadata to get the download URL, then downloads from that URL.
func (c *Client) GetIssueAttachmentContent(ctx *YouTrackContext, issueID string, attachmentID string) ([]byte, error) {
	attachments, err := c.GetIssueAttachments(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment metadata: %w", err)
	}

	var downloadURL string
	for _, att := range attachments {
		if att.ID == attachmentID {
			downloadURL = att.URL
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("attachment %s not found on issue %s", attachmentID, issueID)
	}

	return c.DownloadByURL(ctx, downloadURL)
}

// DownloadByURL downloads raw content from a YouTrack URL (absolute or relative to base URL).
func (c *Client) DownloadByURL(ctx *YouTrackContext, rawURL string) ([]byte, error) {
	// Resolve relative URLs against base URL
	fullURL := rawURL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		fullURL = c.baseURL + rawURL
	}

	req, err := http.NewRequestWithContext(ctx.Context(), http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ctx.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// AddIssueAttachmentFromBytes uploads content as an attachment to an issue
func (c *Client) AddIssueAttachmentFromBytes(ctx *YouTrackContext, issueID string, content []byte, filename string) (*Attachment, error) {
	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file field
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Write content to form
	_, err = fileWriter.Write(content)
	if err != nil {
		return nil, fmt.Errorf("failed to write content: %w", err)
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

	// Find the newly created attachment
	for _, attachment := range attachments {
		if attachment.Name == filename {
			return attachment, nil
		}
	}

	// Fallback to first attachment
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
