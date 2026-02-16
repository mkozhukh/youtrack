package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkozhukh/youtrack/pkg/youtrack"
)

// UserMatch represents a matched user with match details
type UserMatch struct {
	User      *youtrack.User
	MatchType string // "exact_login", "exact_email", "partial_login", "partial_name", "partial_email"
}

// ResolveUser resolves a user query to a specific user login
// It supports:
// - Exact ID match (if query looks like an ID)
// - Exact login match
// - Exact email match
// - Partial login/name/email match
//
// Returns: (userLogin, error)
// If single match found, returns the login
// If no match or multiple matches, returns a ResolveError with helpful context
func (r *Resolver) ResolveUser(ctx context.Context, projectID, query string) (string, error) {
	if query == "" {
		return "", &ResolveError{
			Field:   "user",
			Query:   query,
			Message: "user query cannot be empty",
		}
	}

	query = strings.TrimSpace(query)

	// Fetch all project users
	allUsers, err := r.fetchAllProjectUsers(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch project users: %w", err)
	}

	if len(allUsers) == 0 {
		return "", &ResolveError{
			Field:      "user",
			Query:      query,
			Message:    fmt.Sprintf("no users found in project '%s'", projectID),
			Suggestion: "Check if the project ID is correct and you have access to it.",
		}
	}

	// Try to find matches
	matches := r.findUserMatches(allUsers, query)

	// Handle results
	switch len(matches) {
	case 0:
		// No matches - provide helpful error with available users
		return "", r.noUserMatchError(query, allUsers, projectID)

	case 1:
		// Single match - success
		return matches[0].User.Login, nil

	default:
		// Multiple matches - provide candidates
		return "", r.multipleUserMatchError(query, matches)
	}
}

// fetchAllProjectUsers fetches all users from a project with pagination
func (r *Resolver) fetchAllProjectUsers(ctx context.Context, projectID string) ([]*youtrack.User, error) {
	var allUsers []*youtrack.User
	skip := 0
	top := 100

	for {
		users, err := r.client.GetProjectUsers(ctx, projectID, skip, top)
		if err != nil {
			return nil, err
		}

		if len(users) == 0 {
			break
		}

		allUsers = append(allUsers, users...)

		if len(users) < top {
			break
		}

		skip += len(users)
	}

	return allUsers, nil
}

// findUserMatches finds all users matching the query
func (r *Resolver) findUserMatches(users []*youtrack.User, query string) []UserMatch {
	var exactMatches []UserMatch
	var partialMatches []UserMatch

	normalizedQuery := normalizeString(query)

	for _, user := range users {
		// Check exact matches first (higher priority)
		if equalsNormalized(user.Login, query) {
			exactMatches = append(exactMatches, UserMatch{User: user, MatchType: "exact_login"})
			continue
		}
		if equalsNormalized(user.Email, query) {
			exactMatches = append(exactMatches, UserMatch{User: user, MatchType: "exact_email"})
			continue
		}
		if equalsNormalized(user.FullName, query) {
			exactMatches = append(exactMatches, UserMatch{User: user, MatchType: "exact_name"})
			continue
		}

		// Check partial matches
		if containsNormalized(user.Login, query) {
			partialMatches = append(partialMatches, UserMatch{User: user, MatchType: "partial_login"})
			continue
		}
		if containsNormalized(user.FullName, query) {
			partialMatches = append(partialMatches, UserMatch{User: user, MatchType: "partial_name"})
			continue
		}
		if containsNormalized(user.Email, query) {
			partialMatches = append(partialMatches, UserMatch{User: user, MatchType: "partial_email"})
			continue
		}

		// Check if query words match parts of full name
		queryWords := strings.Fields(normalizedQuery)
		if len(queryWords) > 1 {
			nameWords := strings.Fields(normalizeString(user.FullName))
			matchCount := 0
			for _, qw := range queryWords {
				for _, nw := range nameWords {
					if strings.HasPrefix(nw, qw) || strings.HasPrefix(qw, nw) {
						matchCount++
						break
					}
				}
			}
			if matchCount == len(queryWords) {
				partialMatches = append(partialMatches, UserMatch{User: user, MatchType: "partial_name"})
			}
		}
	}

	// Prefer exact matches over partial
	if len(exactMatches) > 0 {
		return exactMatches
	}

	return partialMatches
}

// noUserMatchError creates an error for no user match
func (r *Resolver) noUserMatchError(query string, users []*youtrack.User, projectID string) *ResolveError {
	// Show some available users as suggestions
	var candidates []string
	maxCandidates := 5

	for i, user := range users {
		if i >= maxCandidates {
			candidates = append(candidates, fmt.Sprintf("... and %d more", len(users)-maxCandidates))
			break
		}
		if user.FullName != "" {
			candidates = append(candidates, fmt.Sprintf("%s (%s)", user.Login, user.FullName))
		} else {
			candidates = append(candidates, user.Login)
		}
	}

	return &ResolveError{
		Field:      "user",
		Query:      query,
		Message:    fmt.Sprintf("no user found matching '%s'", query),
		Candidates: candidates,
		Suggestion: fmt.Sprintf("Use get_project_users(project_id=\"%s\") to see all available users.", projectID),
	}
}

// multipleUserMatchError creates an error for multiple user matches
func (r *Resolver) multipleUserMatchError(query string, matches []UserMatch) *ResolveError {
	var candidates []string
	for _, m := range matches {
		if m.User.FullName != "" {
			candidates = append(candidates, fmt.Sprintf("%s (%s) - %s", m.User.Login, m.User.FullName, m.MatchType))
		} else {
			candidates = append(candidates, fmt.Sprintf("%s - %s", m.User.Login, m.MatchType))
		}
	}

	return &ResolveError{
		Field:      "user",
		Query:      query,
		Message:    fmt.Sprintf("multiple users match '%s' - please be more specific", query),
		Candidates: candidates,
		Suggestion: "Provide the exact login name to avoid ambiguity.",
	}
}

// FormatUserForDisplay formats a user for display in messages
func FormatUserForDisplay(user *youtrack.User) string {
	if user.FullName != "" {
		return fmt.Sprintf("%s (%s)", user.Login, user.FullName)
	}
	return user.Login
}
