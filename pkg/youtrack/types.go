package youtrack

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// YouTrackTime is a custom time type that handles YouTrack's Unix epoch millisecond timestamps
type YouTrackTime struct {
	time.Time
}

// UnmarshalJSON unmarshals YouTrack timestamp (Unix epoch milliseconds) to time.Time
func (yt *YouTrackTime) UnmarshalJSON(data []byte) error {
	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return fmt.Errorf("failed to unmarshal timestamp: %w", err)
	}

	// Convert milliseconds to time.Time
	yt.Time = time.Unix(0, timestamp*int64(time.Millisecond))
	return nil
}

// MarshalJSON marshals time.Time to YouTrack timestamp (Unix epoch milliseconds)
func (yt YouTrackTime) MarshalJSON() ([]byte, error) {
	milliseconds := yt.Time.UnixNano() / int64(time.Millisecond)
	return json.Marshal(milliseconds)
}

type Issue struct {
	ID          string        `json:"idReadable"`
	Summary     string        `json:"summary"`
	Description string        `json:"description,omitempty"`
	Created     YouTrackTime  `json:"created"`
	Updated     YouTrackTime  `json:"updated"`
	Resolved    *YouTrackTime `json:"resolved,omitempty"`
	Reporter    *User         `json:"reporter,omitempty"`
	UpdatedBy   *User         `json:"updatedBy,omitempty"`
	Assignee    *User         `json:"assignee,omitempty"`
	Tags        []*IssueTag   `json:"tags,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"fullName,omitempty"`
	Email    string `json:"email,omitempty"`
}

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description,omitempty"`
}

type IssueComment struct {
	ID      string       `json:"id"`
	Author  *User        `json:"author,omitempty"`
	Text    string       `json:"text"`
	Created YouTrackTime `json:"created"`
	Updated YouTrackTime `json:"updated"`
}

type ProjectRef struct {
	ID string `json:"shortName"`
}

type CustomField struct {
	Name  string      `json:"name"`
	Type  string      `json:"$type"`
	Value interface{} `json:"value"`
}

type SingleValue struct {
	Value interface{} `json:"name"`
}

type SingleUserValue struct {
	ID string `json:"login"`
}

type CreateIssueRequest struct {
	Project     ProjectRef    `json:"project"`
	Summary     string        `json:"summary"`
	Description string        `json:"description,omitempty"`
	Fields      []CustomField `json:"customFields,omitempty"`
}

type UpdateIssueRequest struct {
	Summary     *string       `json:"summary,omitempty"`
	Description *string       `json:"description,omitempty"`
	Fields      []CustomField `json:"customFields,omitempty"`
}

type SearchIssuesRequest struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields,omitempty"`
	Skip   int      `json:"$skip,omitempty"`
	Top    int      `json:"$top,omitempty"`
}

// YouTrackColor represents a color in YouTrack which can be either a string or an object
type YouTrackColor struct {
	value string
}

// UnmarshalJSON handles both string and object color formats from YouTrack API
func (c *YouTrackColor) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var colorStr string
	if err := json.Unmarshal(data, &colorStr); err == nil {
		c.value = colorStr
		return nil
	}

	// Try to unmarshal as object with various possible fields
	var colorObj map[string]interface{}
	if err := json.Unmarshal(data, &colorObj); err != nil {
		return fmt.Errorf("failed to unmarshal color: %w", err)
	}

	// Extract color value from object - YouTrack might use different field names
	if id, ok := colorObj["id"].(string); ok {
		c.value = id
	} else if name, ok := colorObj["name"].(string); ok {
		c.value = name
	} else if bg, ok := colorObj["bg"].(string); ok {
		c.value = bg
	} else if fg, ok := colorObj["fg"].(string); ok {
		c.value = fg
	} else if background, ok := colorObj["background"].(string); ok {
		c.value = background
	} else {
		// If no known field is found, use empty string
		c.value = ""
	}

	return nil
}

// MarshalJSON marshals the color as a string
func (c YouTrackColor) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

// String returns the color as a string
func (c YouTrackColor) String() string {
	return c.value
}

// IsEmpty returns true if the color is empty
func (c YouTrackColor) IsEmpty() bool {
	return c.value == ""
}

type Tag struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Color YouTrackColor `json:"color,omitempty"`
}

type IssueTag struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Color YouTrackColor `json:"color,omitempty"`
}

// DurationValue represents a YouTrack duration period value
type DurationValue struct {
	Minutes      int    `json:"minutes"`
	Presentation string `json:"presentation,omitempty"`
}

// ParseDuration parses a human-friendly duration string into minutes.
// Supported units: w (weeks, 5 work days), d (days, 8 hours), h (hours), m (minutes).
// Examples: "2d", "1h 30m", "1w 2d", "90m", "120" (plain number = minutes).
func ParseDuration(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	re := regexp.MustCompile(`(\d+)\s*([wdhm]?)`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format (examples: '1h', '30m', '2d', '1w')")
	}

	totalMinutes := 0
	usedUnits := make(map[string]bool)

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		valueStr := match[1]
		unit := match[2]
		if unit == "" {
			unit = "m"
		}
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
		case "w":
			totalMinutes += value * 5 * 8 * 60
		case "d":
			totalMinutes += value * 8 * 60
		case "h":
			totalMinutes += value * 60
		case "m":
			totalMinutes += value
		default:
			return 0, fmt.Errorf("invalid time unit '%s' (use 'w' for weeks, 'd' for days, 'h' for hours, or 'm' for minutes)", unit)
		}
	}

	if totalMinutes <= 0 {
		return 0, fmt.Errorf("duration must be greater than 0")
	}

	return totalMinutes, nil
}

type WorkItem struct {
	ID          string        `json:"id"`
	Author      *User         `json:"author,omitempty"`
	Date        YouTrackTime  `json:"date"`
	Duration    DurationValue `json:"duration"`
	Description string        `json:"text,omitempty"`
	Type        *WorkType     `json:"type,omitempty"`
	Issue       *Issue        `json:"issue,omitempty"`
}

type WorkType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Attachment struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Size     int64        `json:"size"`
	Created  YouTrackTime `json:"created"`
	Author   *User        `json:"author,omitempty"`
	MimeType string       `json:"mimeType,omitempty"`
	URL      string       `json:"url,omitempty"`
}

type CreateWorklogRequest struct {
	Duration    DurationValue    `json:"duration"`
	Description string           `json:"text,omitempty"`
	Date        *int64           `json:"date,omitempty"` // Unix epoch milliseconds
	Type        *WorkTypeRequest `json:"type,omitempty"`
}

type WorkTypeRequest struct {
	Name string `json:"name"`
}

type IssueLink struct {
	ID        string    `json:"id"`
	Direction string    `json:"direction"`
	LinkType  *LinkType `json:"linkType,omitempty"`
	Issues    []*Issue  `json:"issues,omitempty"`
}

type LinkType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateIssueLinkRequest struct {
	Query  string      `json:"query"`
	Issues []*IssueRef `json:"issues"`
}

type IssueRef struct {
	ID string `json:"idReadable"`
}

type CustomFieldValue struct {
	Name  string      `json:"name"`
	Type  string      `json:"$type"`
	Value interface{} `json:"value"`
}

type AllowedValue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CommandRequest struct {
	Query  string      `json:"query"`
	Issues []*IssueRef `json:"issues"`
}

// ActivityItem represents an activity item in the issue history
type ActivityItem struct {
	ID            string        `json:"id"`
	Category      Category      `json:"category"`
	Author        *User         `json:"author,omitempty"`
	Timestamp     YouTrackTime  `json:"timestamp"`
	TargetMember  string        `json:"targetMember,omitempty"`
	Field         *Field        `json:"field,omitempty"`
	RemovedValues []*FieldValue `json:"removed,omitempty"`
	AddedValues   []*FieldValue `json:"added,omitempty"`
	Added         *FieldValue   `json:"added,omitempty"`
	Removed       *FieldValue   `json:"removed,omitempty"`
}

// Category represents the category of an activity item
type Category struct {
	ID string `json:"id"`
}

// Field represents a field in an activity item
type Field struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// FieldValue represents a field value in an activity item
type FieldValue struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Text     string `json:"text,omitempty"`
	FullName string `json:"fullName,omitempty"`
	Login    string `json:"login,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}
