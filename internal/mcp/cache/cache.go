package cache

import (
	"sync"
	"time"

	"github.com/mkozhukh/youtrack/pkg/youtrack"
)

// entry holds a cached value with expiration time
type entry struct {
	value      interface{}
	expiration time.Time
}

// isExpired checks if the entry has expired
func (e *entry) isExpired() bool {
	return time.Now().After(e.expiration)
}

// ProjectCache provides per-project caching for metadata
type ProjectCache struct {
	mu           sync.RWMutex
	ttl          time.Duration
	customFields map[string]*entry // projectID -> custom fields
	users        map[string]*entry // projectID -> users
}

// NewProjectCache creates a new cache with the specified TTL
func NewProjectCache(ttl time.Duration) *ProjectCache {
	return &ProjectCache{
		ttl:          ttl,
		customFields: make(map[string]*entry),
		users:        make(map[string]*entry),
	}
}

// GetCustomFields retrieves cached custom fields for a project
// Returns nil if not cached or expired
func (c *ProjectCache) GetCustomFields(projectID string) []*youtrack.CustomField {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.customFields[projectID]
	if !ok || e.isExpired() {
		return nil
	}

	return e.value.([]*youtrack.CustomField)
}

// SetCustomFields stores custom fields for a project
func (c *ProjectCache) SetCustomFields(projectID string, fields []*youtrack.CustomField) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.customFields[projectID] = &entry{
		value:      fields,
		expiration: time.Now().Add(c.ttl),
	}
}

// GetUsers retrieves cached users for a project
// Returns nil if not cached or expired
func (c *ProjectCache) GetUsers(projectID string) []*youtrack.User {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.users[projectID]
	if !ok || e.isExpired() {
		return nil
	}

	return e.value.([]*youtrack.User)
}

// SetUsers stores users for a project
func (c *ProjectCache) SetUsers(projectID string, users []*youtrack.User) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.users[projectID] = &entry{
		value:      users,
		expiration: time.Now().Add(c.ttl),
	}
}

// DropProject removes all cached data for a specific project
func (c *ProjectCache) DropProject(projectID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.customFields, projectID)
	delete(c.users, projectID)
}

// DropAll clears all cached data
func (c *ProjectCache) DropAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.customFields = make(map[string]*entry)
	c.users = make(map[string]*entry)
}
