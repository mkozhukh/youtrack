package tickets

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mkozhukh/youtrack/pkg/youtrack"

	"github.com/spf13/cobra"
)

// isValidTicketID validates the ticket ID format (e.g., "PRJ-123")
func isValidTicketID(ticketID string) bool {
	// Pattern: one or more alphanumeric characters, dash, one or more digits
	pattern := `^[A-Za-z0-9]+-\d+$`
	matched, err := regexp.MatchString(pattern, ticketID)
	return err == nil && matched
}

// getOutputFlag gets the output flag from the command hierarchy
func getOutputFlag(cmd *cobra.Command) string {
	// Walk up the command hierarchy to find the output flag
	current := cmd
	for current != nil {
		if flag := current.Flag("output"); flag != nil {
			return flag.Value.String()
		}
		current = current.Parent()
	}
	return "text" // default
}

// outputResult outputs data in the requested format (text or JSON)
func outputResult(cmd *cobra.Command, data interface{}, formatAsText func(interface{}) error) error {
	outputFlag := getOutputFlag(cmd)

	switch outputFlag {
	case "json":
		return outputJSON(data)
	case "text":
		return formatAsText(data)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlag)
	}
}

// outputJSON outputs data as JSON
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// parseCustomFields parses key=value pairs from the --field flags
func parseCustomFields(fields []string) ([]youtrack.CustomField, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	customFields := make([]youtrack.CustomField, 0, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field format: %s (expected key=value)", field)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty field key in: %s", field)
		}

		if strings.Contains(key, "|") {
			parts := strings.SplitN(key, "|", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid field format: %s (expected key|value)", field)
			}
			key = parts[0]
			keyType := parts[1]

			switch keyType {
			case "SingleEnumIssueCustomField":
			case "enum":
				customFields = append(customFields, youtrack.CustomField{
					Name:  key,
					Type:  "SingleEnumIssueCustomField",
					Value: youtrack.SingleValue{Value: value},
				})
			default:
				customFields = append(customFields, youtrack.CustomField{
					Name:  key,
					Type:  keyType,
					Value: value,
				})
			}

			continue
		}

		customFields = append(customFields, youtrack.CustomField{
			Name:  key,
			Type:  "SimpleIssueCustomField",
			Value: value,
		})
	}

	return customFields, nil
}

// parseDuration parses a duration string like "1h 30m" or "90m" into minutes
func parseDuration(durationStr string) (int, error) {
	return youtrack.ParseDuration(durationStr)
}

// formatDuration formats minutes into a human-readable duration
func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if remainingMinutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
}

// formatFileSize formats file size in a human-readable way
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
