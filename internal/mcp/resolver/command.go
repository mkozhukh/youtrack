package resolver

import (
	"context"
	"regexp"
	"strings"
)

// CommandField represents a parsed field-value pair from a command
type CommandField struct {
	Field    string
	Value    string
	Original string // Original text in command
}

// Common command field patterns
var commandPatterns = []struct {
	Pattern   *regexp.Regexp
	FieldName string
	IsUser    bool
}{
	{regexp.MustCompile(`(?i)^(State)\s+(.+)$`), "State", false},
	{regexp.MustCompile(`(?i)^(Priority)\s+(.+)$`), "Priority", false},
	{regexp.MustCompile(`(?i)^(Type)\s+(.+)$`), "Type", false},
	{regexp.MustCompile(`(?i)^(Assignee)\s+(.+)$`), "Assignee", true},
	{regexp.MustCompile(`(?i)^(Subsystem)\s+(.+)$`), "Subsystem", false},
	{regexp.MustCompile(`(?i)^(Severity)\s+(.+)$`), "Severity", false},
	{regexp.MustCompile(`(?i)^(Resolution)\s+(.+)$`), "Resolution", false},
}

// ParseCommand parses a YouTrack command string into field-value pairs
// Supports commands like: "State Fixed", "Priority Critical", "Assignee john"
func ParseCommand(command string) []CommandField {
	var fields []CommandField

	// Split by common delimiters (space followed by uppercase or known field names)
	// For now, handle single field commands
	command = strings.TrimSpace(command)

	for _, pattern := range commandPatterns {
		if matches := pattern.Pattern.FindStringSubmatch(command); matches != nil {
			fields = append(fields, CommandField{
				Field:    pattern.FieldName,
				Value:    strings.TrimSpace(matches[2]),
				Original: command,
			})
			return fields
		}
	}

	// If no pattern matched, return the command as-is (might be a tag or other command)
	return fields
}

// ResolveCommand resolves field values in a command using smart matching
// Returns the command with resolved values, or an error if resolution fails
func (r *Resolver) ResolveCommand(ctx context.Context, projectID, command string) (string, error) {
	fields := ParseCommand(command)

	if len(fields) == 0 {
		// No known fields to resolve, return as-is
		return command, nil
	}

	resolvedCommand := command

	for _, field := range fields {
		var resolvedValue string
		var err error

		// Check if this is a user field
		isUserField := false
		for _, pattern := range commandPatterns {
			if strings.EqualFold(pattern.FieldName, field.Field) && pattern.IsUser {
				isUserField = true
				break
			}
		}

		if isUserField {
			// Resolve as user
			resolvedValue, err = r.ResolveUser(ctx, projectID, field.Value)
		} else {
			// Resolve as enum value
			resolvedValue, err = r.ResolveEnumValue(ctx, projectID, field.Field, field.Value)
		}

		if err != nil {
			return "", err
		}

		// Replace the value in the command
		// Handle case where the field name case might differ
		for _, pattern := range commandPatterns {
			if strings.EqualFold(pattern.FieldName, field.Field) {
				if matches := pattern.Pattern.FindStringSubmatch(resolvedCommand); matches != nil {
					resolvedCommand = matches[1] + " " + resolvedValue
				}
				break
			}
		}
	}

	return resolvedCommand, nil
}

// IsResolvableCommand checks if a command contains fields that can be resolved
func IsResolvableCommand(command string) bool {
	fields := ParseCommand(command)
	return len(fields) > 0
}
