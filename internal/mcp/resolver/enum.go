package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkozhukh/youtrack/pkg/youtrack"
)

// EnumMatch represents a matched enum value with match details
type EnumMatch struct {
	Value     youtrack.AllowedValue
	MatchType string // "exact", "exact_case_insensitive", "partial", "prefix"
}

// ResolveEnumValue resolves an enum field value query to a specific value
// It supports:
// - Exact match
// - Case-insensitive exact match
// - Prefix match
// - Partial (substring) match
//
// Returns: (resolvedValue, error)
// If single match found, returns the exact value name
// If no match or multiple matches, returns a ResolveError with helpful context
func (r *Resolver) ResolveEnumValue(ctx context.Context, projectID, fieldName, query string) (string, error) {
	if query == "" {
		return "", &ResolveError{
			Field:   fieldName,
			Query:   query,
			Message: fmt.Sprintf("%s value cannot be empty", fieldName),
		}
	}

	query = strings.TrimSpace(query)

	// Fetch allowed values for this field
	allowedValues, err := r.client.GetCustomFieldAllowedValues(ctx, projectID, fieldName)
	if err != nil {
		// Field might not exist or not be an enum - return the original value
		// The API will handle validation
		return query, nil
	}

	if len(allowedValues) == 0 {
		// No allowed values defined - return original value
		return query, nil
	}

	// Try to find matches
	matches := r.findEnumMatches(allowedValues, query)

	// Handle results
	switch len(matches) {
	case 0:
		// No matches - provide helpful error with available values
		return "", r.noEnumMatchError(fieldName, query, allowedValues)

	case 1:
		// Single match - success
		return matches[0].Value.Name, nil

	default:
		// Multiple matches - provide candidates
		return "", r.multipleEnumMatchError(fieldName, query, matches)
	}
}

// findEnumMatches finds all enum values matching the query
func (r *Resolver) findEnumMatches(values []youtrack.AllowedValue, query string) []EnumMatch {
	var exactMatches []EnumMatch
	var prefixMatches []EnumMatch
	var partialMatches []EnumMatch

	normalizedQuery := normalizeString(query)

	for _, value := range values {
		normalizedValue := normalizeString(value.Name)

		// Exact match (case-sensitive)
		if value.Name == query {
			exactMatches = append(exactMatches, EnumMatch{Value: value, MatchType: "exact"})
			continue
		}

		// Exact match (case-insensitive)
		if normalizedValue == normalizedQuery {
			exactMatches = append(exactMatches, EnumMatch{Value: value, MatchType: "exact_case_insensitive"})
			continue
		}

		// Prefix match (value starts with query)
		if strings.HasPrefix(normalizedValue, normalizedQuery) {
			prefixMatches = append(prefixMatches, EnumMatch{Value: value, MatchType: "prefix"})
			continue
		}

		// Partial match (value contains query)
		if strings.Contains(normalizedValue, normalizedQuery) {
			partialMatches = append(partialMatches, EnumMatch{Value: value, MatchType: "partial"})
			continue
		}

		// Word-based matching for multi-word values
		// e.g., "progress" matches "In Progress"
		valueWords := strings.Fields(normalizedValue)
		for _, word := range valueWords {
			if strings.HasPrefix(word, normalizedQuery) {
				partialMatches = append(partialMatches, EnumMatch{Value: value, MatchType: "partial_word"})
				break
			}
		}
	}

	// Return matches by priority: exact > prefix > partial
	if len(exactMatches) > 0 {
		return exactMatches
	}
	if len(prefixMatches) > 0 {
		return prefixMatches
	}

	return partialMatches
}

// noEnumMatchError creates an error for no enum value match
func (r *Resolver) noEnumMatchError(fieldName, query string, values []youtrack.AllowedValue) *ResolveError {
	var candidates []string
	for _, v := range values {
		candidates = append(candidates, v.Name)
	}

	return &ResolveError{
		Field:      fieldName,
		Query:      query,
		Message:    fmt.Sprintf("'%s' is not a valid %s value", query, fieldName),
		Candidates: candidates,
		Suggestion: fmt.Sprintf("Use one of the listed values for %s.", fieldName),
	}
}

// multipleEnumMatchError creates an error for multiple enum value matches
func (r *Resolver) multipleEnumMatchError(fieldName, query string, matches []EnumMatch) *ResolveError {
	var candidates []string
	for _, m := range matches {
		candidates = append(candidates, fmt.Sprintf("%s (%s)", m.Value.Name, m.MatchType))
	}

	return &ResolveError{
		Field:      fieldName,
		Query:      query,
		Message:    fmt.Sprintf("multiple %s values match '%s' - please be more specific", fieldName, query),
		Candidates: candidates,
		Suggestion: "Provide a more complete value name to avoid ambiguity.",
	}
}

// Common field names that are typically enums
var CommonEnumFields = []string{
	"State",
	"Priority",
	"Type",
	"Subsystem",
	"Severity",
	"Resolution",
}

// IsCommonEnumField checks if a field name is a commonly used enum field
func IsCommonEnumField(fieldName string) bool {
	normalizedField := normalizeString(fieldName)
	for _, f := range CommonEnumFields {
		if normalizeString(f) == normalizedField {
			return true
		}
	}
	return false
}
