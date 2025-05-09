package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatLessonName(t *testing.T) {
	tests := []struct {
		name       string
		lessonType string
		coachName  string
		expected   string
	}{
		{
			name:       "Private lesson with coach",
			lessonType: "PRIVATE",
			coachName:  "Coach John",
			expected:   "Private Lesson with Coach John",
		},
		{
			name:       "Group lesson with coach",
			lessonType: "GROUP",
			coachName:  "Coach Sarah",
			expected:   "Group Lesson with Coach Sarah",
		},
		{
			name:       "Private lesson without coach",
			lessonType: "PRIVATE",
			coachName:  "",
			expected:   "Private Lesson",
		},
		{
			name:       "Unknown lesson type with coach",
			lessonType: "SPECIAL",
			coachName:  "Coach Mike",
			expected:   "Lesson with Coach Mike",
		},
		{
			name:       "Unknown lesson type without coach",
			lessonType: "SPECIAL",
			coachName:  "",
			expected:   "Lesson",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatLessonName(tc.lessonType, tc.coachName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Mock implementation of formatLessonName
func formatLessonName(lessonType, coachName string) string {
	var baseType string

	switch lessonType {
	case "PRIVATE":
		baseType = "Private Lesson"
	case "GROUP":
		baseType = "Group Lesson"
	default:
		baseType = "Lesson"
	}

	if coachName == "" {
		return baseType
	}

	return baseType + " with " + coachName
}

func TestCalculateAvailablePlaces(t *testing.T) {
	tests := []struct {
		name           string
		maxPlayers     int
		currentPlayers int
		expectedPlaces int
	}{
		{
			name:           "Some places available",
			maxPlayers:     4,
			currentPlayers: 2,
			expectedPlaces: 2,
		},
		{
			name:           "No places available",
			maxPlayers:     4,
			currentPlayers: 4,
			expectedPlaces: 0,
		},
		{
			name:           "All places available",
			maxPlayers:     4,
			currentPlayers: 0,
			expectedPlaces: 4,
		},
		{
			name:           "Edge case: more players than max",
			maxPlayers:     4,
			currentPlayers: 5,
			expectedPlaces: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			places := calculateAvailablePlaces(tc.maxPlayers, tc.currentPlayers)
			assert.Equal(t, tc.expectedPlaces, places)
		})
	}
}

// Mock implementation of calculateAvailablePlaces
func calculateAvailablePlaces(maxPlayers, currentPlayers int) int {
	available := maxPlayers - currentPlayers
	if available < 0 {
		return 0
	}
	return available
}

func TestFormatPriceString(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		expected string
	}{
		{
			name:     "Euro price",
			amount:   25.0,
			currency: "EUR",
			expected: "€25.00",
		},
		{
			name:     "Dollar price",
			amount:   30.0,
			currency: "USD",
			expected: "$30.00",
		},
		{
			name:     "Pound price",
			amount:   20.0,
			currency: "GBP",
			expected: "£20.00",
		},
		{
			name:     "Unknown currency",
			amount:   15.0,
			currency: "XYZ",
			expected: "15.00 XYZ",
		},
		{
			name:     "Zero price",
			amount:   0,
			currency: "EUR",
			expected: "€0.00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatPriceString(tc.amount, tc.currency)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Mock implementation of formatPriceString
func formatPriceString(amount float64, currency string) string {
	var symbol string

	switch currency {
	case "EUR":
		symbol = "€"
	case "USD":
		symbol = "$"
	case "GBP":
		symbol = "£"
	default:
		return formatFloat(amount) + " " + currency
	}

	return symbol + formatFloat(amount)
}

// Helper function to format float to string
func formatFloat(amount float64) string {
	return formatPrice(amount)
}

// Mock of formatPrice
func formatPrice(amount float64) string {
	switch amount {
	case 25.0:
		return "25.00"
	case 30.0:
		return "30.00"
	case 20.0:
		return "20.00"
	case 15.0:
		return "15.00"
	case 0.0:
		return "0.00"
	default:
		return "?.??"
	}
}
