package tracker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"

	"github.com/charmbracelet/log"
)

// APIKeyProvider provides the effective API key for a context
type APIKeyProvider interface {
	GetEffectiveAPIKey(ctx context.Context) string
}

// ProjectTracker tracks the last used project per user (identified by auth key hash)
type ProjectTracker struct {
	mu       sync.RWMutex
	projects map[string]string // keyHash -> projectID
	filePath string
}

// NewProjectTracker creates a new tracker, loading state from file if it exists
func NewProjectTracker(filePath string) *ProjectTracker {
	pt := &ProjectTracker{
		projects: make(map[string]string),
		filePath: filePath,
	}
	pt.load()
	return pt
}

// HashKey returns a SHA256 hash of the API key for use as user identifier
func HashKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// GetLastProject returns the last used project for the given key hash
func (pt *ProjectTracker) GetLastProject(keyHash string) string {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.projects[keyHash]
}

// SetLastProject sets the last used project for the given key hash
func (pt *ProjectTracker) SetLastProject(keyHash, projectID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.projects[keyHash] == projectID {
		return // no change
	}

	pt.projects[keyHash] = projectID
	pt.save()
}

// load reads the tracker state from file
func (pt *ProjectTracker) load() {
	if pt.filePath == "" {
		return
	}

	data, err := os.ReadFile(pt.filePath)
	if err != nil {
		return // file doesn't exist or can't be read, start fresh
	}

	if err := json.Unmarshal(data, &pt.projects); err != nil {
		log.Error("Failed to parse project tracker file", "path", pt.filePath, "error", err)
	}
}

// save writes the tracker state to file
func (pt *ProjectTracker) save() {
	if pt.filePath == "" {
		return
	}

	data, err := json.MarshalIndent(pt.projects, "", "  ")
	if err != nil {
		log.Error("Failed to marshal project tracker data", "error", err)
		return
	}

	if err := os.WriteFile(pt.filePath, data, 0600); err != nil {
		log.Error("Failed to save project tracker file", "path", pt.filePath, "error", err)
	}
}

// ContextProjectTracker wraps ProjectTracker with context-aware API key resolution
type ContextProjectTracker struct {
	tracker     *ProjectTracker
	keyProvider APIKeyProvider
}

// NewContextProjectTracker creates a context-aware project tracker
func NewContextProjectTracker(tracker *ProjectTracker, keyProvider APIKeyProvider) *ContextProjectTracker {
	return &ContextProjectTracker{
		tracker:     tracker,
		keyProvider: keyProvider,
	}
}

// TrackProject records the project as last used for the current user
func (ct *ContextProjectTracker) TrackProject(ctx context.Context, projectID string) {
	if projectID == "" {
		return
	}
	apiKey := ct.keyProvider.GetEffectiveAPIKey(ctx)
	if apiKey == "" {
		return
	}
	ct.tracker.SetLastProject(HashKey(apiKey), projectID)
}

// GetLastProject returns the last used project for the current user
func (ct *ContextProjectTracker) GetLastProject(ctx context.Context) string {
	apiKey := ct.keyProvider.GetEffectiveAPIKey(ctx)
	if apiKey == "" {
		return ""
	}
	return ct.tracker.GetLastProject(HashKey(apiKey))
}
