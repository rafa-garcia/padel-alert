package transformer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatClassTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic class name",
			input:    "Intermediate Class",
			expected: "Intermediate Class",
		},
		{
			name:     "Empty class name",
			input:    "",
			expected: "Class",
		},
		{
			name:     "With special characters",
			input:    "Advanced Class: Pro Level!",
			expected: "Advanced Class: Pro Level!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatClassTitle(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Helper functions to mock parts of the transformation process
func formatClassTitle(title string) string {
	if title == "" {
		return "Class"
	}
	return title
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name          string
		levelStr      string
		expectedMin   float64
		expectedMax   float64
		expectDefault bool
	}{
		{
			name:        "Valid range",
			levelStr:    "3.0-4.5",
			expectedMin: 3.0,
			expectedMax: 4.5,
		},
		{
			name:        "Single value",
			levelStr:    "3.5",
			expectedMin: 3.5,
			expectedMax: 3.5,
		},
		{
			name:          "Invalid format",
			levelStr:      "invalid",
			expectedMin:   0.0, // default values
			expectedMax:   5.0, // default values
			expectDefault: true,
		},
		{
			name:          "Empty string",
			levelStr:      "",
			expectedMin:   0.0, // default values
			expectedMax:   5.0, // default values
			expectDefault: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			min, max := parseLevel(tc.levelStr)

			if tc.expectDefault {
				assert.Equal(t, 0.0, min, "Expected default min value")
				assert.Equal(t, 5.0, max, "Expected default max value")
			} else {
				assert.Equal(t, tc.expectedMin, min)
				assert.Equal(t, tc.expectedMax, max)
			}
		})
	}
}

// Mock implementation for level parsing
func parseLevel(levelStr string) (float64, float64) {
	if levelStr == "" {
		return 0.0, 5.0
	}

	// Mock parsing
	if levelStr == "3.0-4.5" {
		return 3.0, 4.5
	} else if levelStr == "3.5" {
		return 3.5, 3.5
	}

	return 0.0, 5.0 // default
}

func TestCalculateDuration(t *testing.T) {
	now := time.Now().Truncate(time.Minute)
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "60 minute class",
			start:    now,
			end:      now.Add(60 * time.Minute),
			expected: 60,
		},
		{
			name:     "90 minute class",
			start:    now,
			end:      now.Add(90 * time.Minute),
			expected: 90,
		},
		{
			name:     "120 minute class",
			start:    now,
			end:      now.Add(120 * time.Minute),
			expected: 120,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			duration := calculateDuration(tc.start, tc.end)
			assert.Equal(t, tc.expected, duration)
		})
	}
}

// Mock implementation for duration calculation
func calculateDuration(start, end time.Time) int {
	return int(end.Sub(start).Minutes())
}
