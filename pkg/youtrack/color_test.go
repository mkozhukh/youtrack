package youtrack

import (
	"encoding/json"
	"testing"
)

func TestYouTrackColor_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "String color",
			input:       `"#FF0000"`,
			expected:    "#FF0000",
			expectError: false,
		},
		{
			name:        "Object with id field",
			input:       `{"id": "red", "name": "Red"}`,
			expected:    "red",
			expectError: false,
		},
		{
			name:        "Object with name field",
			input:       `{"name": "blue", "description": "Blue color"}`,
			expected:    "blue",
			expectError: false,
		},
		{
			name:        "Object with bg field",
			input:       `{"bg": "#00FF00", "fg": "#FFFFFF"}`,
			expected:    "#00FF00",
			expectError: false,
		},
		{
			name:        "Object with fg field (no bg)",
			input:       `{"fg": "#FFFFFF", "alpha": 0.5}`,
			expected:    "#FFFFFF",
			expectError: false,
		},
		{
			name:        "Object with background field",
			input:       `{"background": "#FFFF00", "transparency": 0.8}`,
			expected:    "#FFFF00",
			expectError: false,
		},
		{
			name:        "Empty object",
			input:       `{}`,
			expected:    "",
			expectError: false,
		},
		{
			name:        "Object with unknown fields",
			input:       `{"unknown": "value", "another": 123}`,
			expected:    "",
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			input:       `invalid json`,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var color YouTrackColor
			err := json.Unmarshal([]byte(tt.input), &color)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if color.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, color.String())
			}
		})
	}
}

func TestYouTrackColor_MarshalJSON(t *testing.T) {
	color := YouTrackColor{value: "#FF0000"}
	data, err := json.Marshal(color)
	if err != nil {
		t.Fatalf("Failed to marshal color: %v", err)
	}

	expected := `"#FF0000"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestYouTrackColor_Methods(t *testing.T) {
	// Test String method
	color := YouTrackColor{value: "red"}
	if color.String() != "red" {
		t.Errorf("Expected String() to return 'red', got %q", color.String())
	}

	// Test IsEmpty method
	emptyColor := YouTrackColor{value: ""}
	if !emptyColor.IsEmpty() {
		t.Errorf("Expected IsEmpty() to return true for empty color")
	}

	nonEmptyColor := YouTrackColor{value: "blue"}
	if nonEmptyColor.IsEmpty() {
		t.Errorf("Expected IsEmpty() to return false for non-empty color")
	}
}

func TestIssueTag_JSONWithColor(t *testing.T) {
	// Test with string color
	tagJSONString := `{
		"id": "tag-1",
		"name": "Bug",
		"color": "#FF0000"
	}`

	var tag IssueTag
	err := json.Unmarshal([]byte(tagJSONString), &tag)
	if err != nil {
		t.Fatalf("Failed to unmarshal tag with string color: %v", err)
	}

	if tag.Color.String() != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got %q", tag.Color.String())
	}

	// Test with object color
	tagJSONObject := `{
		"id": "tag-2",
		"name": "Feature",
		"color": {"id": "green", "name": "Green"}
	}`

	var tag2 IssueTag
	err = json.Unmarshal([]byte(tagJSONObject), &tag2)
	if err != nil {
		t.Fatalf("Failed to unmarshal tag with object color: %v", err)
	}

	if tag2.Color.String() != "green" {
		t.Errorf("Expected color 'green', got %q", tag2.Color.String())
	}
}

func TestTag_JSONRoundTrip(t *testing.T) {
	original := Tag{
		ID:    "tag-1",
		Name:  "Test Tag",
		Color: YouTrackColor{value: "#ABCDEF"},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal tag: %v", err)
	}

	// Unmarshal back
	var restored Tag
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal tag: %v", err)
	}

	// Verify values
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: expected %q, got %q", original.ID, restored.ID)
	}
	if restored.Name != original.Name {
		t.Errorf("Name mismatch: expected %q, got %q", original.Name, restored.Name)
	}
	if restored.Color.String() != original.Color.String() {
		t.Errorf("Color mismatch: expected %q, got %q", original.Color.String(), restored.Color.String())
	}
}
