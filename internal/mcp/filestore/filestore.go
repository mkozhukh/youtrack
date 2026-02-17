package filestore

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type entry struct {
	path        string
	filename    string
	contentType string
	expiresAt   time.Time
}

// Store manages temporary files with TTL-based expiry.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*entry
	tempDir string
	ttl     time.Duration
	maxSize int64
	done    chan struct{}
}

// NewStore creates a new file store. It creates a temp directory and starts a
// background goroutine that cleans up expired entries every 60 seconds.
func NewStore(ttl time.Duration, maxSizeMB int) (*Store, error) {
	dir, err := os.MkdirTemp("", "youtrack-filestore-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	s := &Store{
		entries: make(map[string]*entry),
		tempDir: dir,
		ttl:     ttl,
		maxSize: int64(maxSizeMB) * 1024 * 1024,
		done:    make(chan struct{}),
	}

	go s.cleanupLoop()
	return s, nil
}

// Put stores data as a temporary file and returns a unique file ID.
func (s *Store) Put(data []byte, filename string) (string, error) {
	if int64(len(data)) > s.maxSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d bytes)", len(data), s.maxSize)
	}

	id := newUUID()
	path := filepath.Join(s.tempDir, id)

	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	ct := mime.TypeByExtension(filepath.Ext(filename))
	if ct == "" {
		ct = "application/octet-stream"
	}

	s.mu.Lock()
	s.entries[id] = &entry{
		path:        path,
		filename:    filename,
		contentType: ct,
		expiresAt:   time.Now().Add(s.ttl),
	}
	s.mu.Unlock()

	return id, nil
}

// PutFromFile stores a file from a given path and returns a unique file ID.
func (s *Store) PutFromFile(srcPath, filename string) (string, error) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	if info.Size() > s.maxSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d bytes)", info.Size(), s.maxSize)
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return s.Put(data, filename)
}

// Get returns the file path, original filename, and content type for a stored file.
func (s *Store) Get(fileID string) (filePath, filename, contentType string, err error) {
	s.mu.RLock()
	e, ok := s.entries[fileID]
	s.mu.RUnlock()

	if !ok {
		return "", "", "", fmt.Errorf("file not found: %s", fileID)
	}

	if time.Now().After(e.expiresAt) {
		s.remove(fileID)
		return "", "", "", fmt.Errorf("file expired: %s", fileID)
	}

	return e.path, e.filename, e.contentType, nil
}

// Close stops the cleanup goroutine and removes the temp directory.
func (s *Store) Close() {
	close(s.done)
	os.RemoveAll(s.tempDir)
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

func (s *Store) cleanup() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, e := range s.entries {
		if now.After(e.expiresAt) {
			os.Remove(e.path)
			delete(s.entries, id)
		}
	}
}

func (s *Store) remove(fileID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.entries[fileID]; ok {
		os.Remove(e.path)
		delete(s.entries, fileID)
	}
}

func newUUID() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
