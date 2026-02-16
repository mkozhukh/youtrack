package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CacheManager defines the interface for cache management operations
type CacheManager interface {
	DropProject(projectID string)
	DropAll()
}

// CacheHandlers manages cache-related MCP operations
type CacheHandlers struct {
	cache      CacheManager
	toolLogger func(string, map[string]interface{})
}

// NewCacheHandlers creates a new instance of CacheHandlers
func NewCacheHandlers(cache CacheManager, toolLogger func(string, map[string]interface{})) *CacheHandlers {
	return &CacheHandlers{
		cache:      cache,
		toolLogger: toolLogger,
	}
}

// DropCacheHandler handles the drop_cache tool call
func (h *CacheHandlers) DropCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectID, _ := args["project_id"].(string)

	if h.toolLogger != nil {
		h.toolLogger("drop_cache", map[string]interface{}{
			"project_id": projectID,
		})
	}

	if projectID != "" {
		h.cache.DropProject(projectID)
		return mcp.NewToolResultText(fmt.Sprintf("Cache dropped for project '%s'.", projectID)), nil
	}

	h.cache.DropAll()
	return mcp.NewToolResultText("Cache dropped for all projects."), nil
}
