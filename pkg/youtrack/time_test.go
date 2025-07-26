package youtrack

import (
	"encoding/json"
	"testing"
	"time"
)

func TestYouTrackTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Time
		expectError bool
	}{
		{
			name:        "Valid timestamp",
			input:       "1640995200000", // 2022-01-01 00:00:00 UTC
			expected:    time.Unix(1640995200, 0),
			expectError: false,
		},
		{
			name:        "Zero timestamp",
			input:       "0",
			expected:    time.Unix(0, 0),
			expectError: false,
		},
		{
			name:        "Invalid input",
			input:       "\"not-a-number\"",
			expected:    time.Time{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var yt YouTrackTime
			err := json.Unmarshal([]byte(tt.input), &yt)

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

			if !yt.Time.Equal(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, yt.Time)
			}
		})
	}
}

func TestYouTrackTime_MarshalJSON(t *testing.T) {
	// Test marshaling
	yt := YouTrackTime{Time: time.Unix(1640995200, 0)} // 2022-01-01 00:00:00 UTC
	data, err := json.Marshal(yt)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := "1640995200000"
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestIssue_JSONRoundTrip(t *testing.T) {
	// Test that Issue struct can properly unmarshal YouTrack API response
	issueJSON := `{
		"id": "TEST-1",
		"summary": "Test Issue",
		"description": "Test description",
		"created": 1640995200000,
		"updated": 1640995260000
	}`

	var issue Issue
	err := json.Unmarshal([]byte(issueJSON), &issue)
	if err != nil {
		t.Fatalf("Failed to unmarshal issue: %v", err)
	}

	expectedCreated := time.Unix(1640995200, 0)
	expectedUpdated := time.Unix(1640995260, 0)

	if !issue.Created.Time.Equal(expectedCreated) {
		t.Errorf("Expected created time %v, got %v", expectedCreated, issue.Created.Time)
	}

	if !issue.Updated.Time.Equal(expectedUpdated) {
		t.Errorf("Expected updated time %v, got %v", expectedUpdated, issue.Updated.Time)
	}

	// Test marshaling back
	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Failed to marshal issue: %v", err)
	}

	// Verify the timestamps are preserved
	var roundTrip Issue
	err = json.Unmarshal(data, &roundTrip)
	if err != nil {
		t.Fatalf("Failed to unmarshal round trip: %v", err)
	}

	if !roundTrip.Created.Time.Equal(issue.Created.Time) {
		t.Errorf("Round trip failed for created time")
	}

	if !roundTrip.Updated.Time.Equal(issue.Updated.Time) {
		t.Errorf("Round trip failed for updated time")
	}
}
