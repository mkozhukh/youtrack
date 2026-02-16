package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkozhukh/youtrack/pkg/youtrack"
)

// ResolveError represents a resolution error with helpful context
type ResolveError struct {
	Field      string
	Query      string
	Message    string
	Candidates []string
	Suggestion string
}

func (e *ResolveError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s", e.Field, e.Message))

	if len(e.Candidates) > 0 {
		sb.WriteString("\n\nDid you mean one of these?\n")
		for _, c := range e.Candidates {
			sb.WriteString(fmt.Sprintf("  - %s\n", c))
		}
	}

	if e.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("\n%s", e.Suggestion))
	}

	return sb.String()
}

// Resolver provides smart resolution for users and enum fields
type Resolver struct {
	client ResolverClient
}

// ResolverClient defines the interface for YouTrack operations needed by resolver
type ResolverClient interface {
	GetProjectUsers(ctx context.Context, projectID string, skip, top int) ([]*youtrack.User, error)
	GetCustomFieldAllowedValues(ctx context.Context, projectID string, fieldName string) ([]youtrack.AllowedValue, error)
}

// NewResolver creates a new resolver instance
func NewResolver(client ResolverClient) *Resolver {
	return &Resolver{client: client}
}

// normalizeString normalizes a string for comparison (lowercase, trimmed)
func normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// containsNormalized checks if haystack contains needle (case-insensitive)
func containsNormalized(haystack, needle string) bool {
	return strings.Contains(normalizeString(haystack), normalizeString(needle))
}

// equalsNormalized checks if two strings are equal (case-insensitive)
func equalsNormalized(a, b string) bool {
	return normalizeString(a) == normalizeString(b)
}
