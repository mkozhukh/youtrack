package tickets

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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
func parseCustomFields(fields []string) (map[string]interface{}, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	customFields := make(map[string]interface{})
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

		// For now, treat all custom field values as strings
		// More complex field types can be handled later
		customFields[key] = value
	}

	return customFields, nil
}

// parseDuration parses a duration string like "1h 30m" or "90m" into minutes
func parseDuration(durationStr string) (int, error) {
	durationStr = strings.TrimSpace(durationStr)
	if durationStr == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	// Regular expression to match time units
	re := regexp.MustCompile(`(\d+)\s*([hm]?)`)
	matches := re.FindAllStringSubmatch(durationStr, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format (examples: '1h', '30m', '1h 30m', '90m')")
	}

	totalMinutes := 0
	usedUnits := make(map[string]bool)

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		valueStr := match[1]
		unit := match[2]

		// Default to minutes if no unit specified
		if unit == "" {
			unit = "m"
		}

		// Check for duplicate units
		if usedUnits[unit] {
			return 0, fmt.Errorf("duplicate time unit '%s' in duration", unit)
		}
		usedUnits[unit] = true

		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return 0, fmt.Errorf("invalid number '%s' in duration", valueStr)
		}

		if value < 0 {
			return 0, fmt.Errorf("negative values not allowed in duration")
		}

		switch unit {
		case "h":
			totalMinutes += value * 60
		case "m":
			totalMinutes += value
		default:
			return 0, fmt.Errorf("invalid time unit '%s' (use 'h' for hours or 'm' for minutes)", unit)
		}
	}

	if totalMinutes <= 0 {
		return 0, fmt.Errorf("duration must be greater than 0")
	}

	return totalMinutes, nil
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
