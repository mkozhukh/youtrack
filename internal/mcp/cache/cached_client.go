package cache

import (
	"context"

	"github.com/mkozhukh/youtrack/pkg/youtrack"
)

// ProjectClient defines the interface for project-related operations
type ProjectClient interface {
	GetProject(ctx context.Context, projectID string) (*youtrack.Project, error)
	GetProjectByName(ctx context.Context, name string) (*youtrack.Project, error)
	ListProjects(ctx context.Context, skip, top int) ([]*youtrack.Project, error)
	GetProjectCustomFields(ctx context.Context, projectID string) ([]*youtrack.CustomField, error)
	GetCustomFieldAllowedValues(ctx context.Context, projectID string, fieldName string) ([]youtrack.AllowedValue, error)
	GetAvailableLinkTypes(ctx context.Context) ([]*youtrack.LinkType, error)
}

// UserClient defines the interface for user-related operations
type UserClient interface {
	GetCurrentUser(ctx context.Context) (*youtrack.User, error)
	GetProjectUsers(ctx context.Context, projectID string, skip, top int) ([]*youtrack.User, error)
}

// CachedClient wraps a client with caching functionality
type CachedClient struct {
	delegate ProjectAndUserClient
	cache    *ProjectCache
}

// ProjectAndUserClient combines ProjectClient and UserClient interfaces
type ProjectAndUserClient interface {
	ProjectClient
	UserClient
}

// NewCachedClient creates a new cached client wrapper
func NewCachedClient(delegate ProjectAndUserClient, cache *ProjectCache) *CachedClient {
	return &CachedClient{
		delegate: delegate,
		cache:    cache,
	}
}

// GetProject delegates to the underlying client (no caching)
func (c *CachedClient) GetProject(ctx context.Context, projectID string) (*youtrack.Project, error) {
	return c.delegate.GetProject(ctx, projectID)
}

// GetProjectByName delegates to the underlying client (no caching)
func (c *CachedClient) GetProjectByName(ctx context.Context, name string) (*youtrack.Project, error) {
	return c.delegate.GetProjectByName(ctx, name)
}

// ListProjects delegates to the underlying client (no caching)
func (c *CachedClient) ListProjects(ctx context.Context, skip, top int) ([]*youtrack.Project, error) {
	return c.delegate.ListProjects(ctx, skip, top)
}

// GetProjectCustomFields returns cached custom fields or fetches from API
func (c *CachedClient) GetProjectCustomFields(ctx context.Context, projectID string) ([]*youtrack.CustomField, error) {
	// Check cache first
	if cached := c.cache.GetCustomFields(projectID); cached != nil {
		return cached, nil
	}

	// Fetch from API
	fields, err := c.delegate.GetProjectCustomFields(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.SetCustomFields(projectID, fields)
	return fields, nil
}

// GetCustomFieldAllowedValues delegates to the underlying client (no caching)
func (c *CachedClient) GetCustomFieldAllowedValues(ctx context.Context, projectID string, fieldName string) ([]youtrack.AllowedValue, error) {
	return c.delegate.GetCustomFieldAllowedValues(ctx, projectID, fieldName)
}

// GetAvailableLinkTypes delegates to the underlying client (no caching)
func (c *CachedClient) GetAvailableLinkTypes(ctx context.Context) ([]*youtrack.LinkType, error) {
	return c.delegate.GetAvailableLinkTypes(ctx)
}

// GetCurrentUser delegates to the underlying client (no caching)
func (c *CachedClient) GetCurrentUser(ctx context.Context) (*youtrack.User, error) {
	return c.delegate.GetCurrentUser(ctx)
}

// GetProjectUsers returns cached users or fetches all pages from API
func (c *CachedClient) GetProjectUsers(ctx context.Context, projectID string, skip, top int) ([]*youtrack.User, error) {
	// Check cache first
	cachedUsers := c.cache.GetUsers(projectID)
	if cachedUsers != nil {
		// Return requested slice from cached data
		if skip >= len(cachedUsers) {
			return []*youtrack.User{}, nil
		}
		end := skip + top
		if end > len(cachedUsers) {
			end = len(cachedUsers)
		}
		return cachedUsers[skip:end], nil
	}

	// Fetch all users from API (paginated)
	var allUsers []*youtrack.User
	fetchSkip := 0
	fetchTop := 100

	for {
		users, err := c.delegate.GetProjectUsers(ctx, projectID, fetchSkip, fetchTop)
		if err != nil {
			return nil, err
		}

		if len(users) == 0 {
			break
		}

		allUsers = append(allUsers, users...)

		if len(users) < fetchTop {
			break
		}

		fetchSkip += len(users)
	}

	// Store complete list in cache
	c.cache.SetUsers(projectID, allUsers)

	// Return requested slice
	if skip >= len(allUsers) {
		return []*youtrack.User{}, nil
	}
	end := skip + top
	if end > len(allUsers) {
		end = len(allUsers)
	}
	return allUsers[skip:end], nil
}

// Cache returns the underlying cache for management operations
func (c *CachedClient) Cache() *ProjectCache {
	return c.cache
}
