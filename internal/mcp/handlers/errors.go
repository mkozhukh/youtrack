package handlers

import (
	"fmt"
	"strings"

	"mkozhukh/youtrack/pkg/youtrack"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
)

// ErrorHandler provides consistent error handling and formatting
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleError processes an error and returns an appropriate MCP error result
func (e *ErrorHandler) HandleError(err error, operation string) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	// Check if it's a YouTrack API error
	if apiErr, ok := err.(*youtrack.APIError); ok {
		return e.handleAPIError(apiErr, operation)
	}

	// Generic error handling
	message := fmt.Sprintf("Error during %s: %s", operation, err.Error())
	return mcp.NewToolResultError(message)
}

// handleAPIError handles YouTrack-specific API errors
func (e *ErrorHandler) handleAPIError(apiErr *youtrack.APIError, operation string) *mcp.CallToolResult {
	var message string

	switch apiErr.StatusCode {
	case 400:
		message = fmt.Sprintf("Bad request during %s: %s", operation, apiErr.Message)
	case 401:
		message = fmt.Sprintf("Authentication failed during %s. Please check your API key.", operation)
	case 403:
		message = fmt.Sprintf("Permission denied during %s. You don't have the required permissions.", operation)
	case 404:
		message = fmt.Sprintf("Resource not found during %s: %s", operation, apiErr.Message)
	case 409:
		message = fmt.Sprintf("Conflict during %s: %s", operation, apiErr.Message)
	case 429:
		message = fmt.Sprintf("Rate limit exceeded during %s. Please try again later.", operation)
	case 500, 502, 503, 504:
		log.Error("YouTrack error", "text", apiErr.Message)
		message = fmt.Sprintf("YouTrack server error during %s. Please try again later.", operation)
	default:
		message = fmt.Sprintf("API error during %s (status %d): %s", operation, apiErr.StatusCode, apiErr.Message)
	}

	return mcp.NewToolResultError(message)
}

// ValidateRequiredParameter validates that a required parameter is not empty
func (e *ErrorHandler) ValidateRequiredParameter(value, paramName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", paramName)
	}
	return nil
}

// ValidatePositiveNumber validates that a number is positive
func (e *ErrorHandler) ValidatePositiveNumber(value float64, paramName string) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", paramName)
	}
	return nil
}

// FormatValidationError formats a validation error
func (e *ErrorHandler) FormatValidationError(paramName string, err error) *mcp.CallToolResult {
	message := fmt.Sprintf("Invalid %s parameter: %s", paramName, err.Error())
	return mcp.NewToolResultError(message)
}
