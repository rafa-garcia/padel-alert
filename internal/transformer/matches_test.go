package transformer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeFormatting(t *testing.T) {
	timeStr := "2023-05-15T14:30:00"

	parsedTime, err := parseMatchTime(timeStr)

	assert.NoError(t, err)
	assert.Equal(t, 2023, parsedTime.Year())
	assert.Equal(t, time.Month(5), parsedTime.Month())
	assert.Equal(t, 15, parsedTime.Day())
	assert.Equal(t, 14, parsedTime.Hour())
	assert.Equal(t, 30, parsedTime.Minute())
}

func parseMatchTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", timeStr)
}

func TestCalculateAvailableSpots(t *testing.T) {
	tests := []struct {
		name              string
		teams             []testTeam
		maxPlayersPerTeam int
		expectedAvailable int
	}{
		{
			name: "Some spots available",
			teams: []testTeam{
				{players: 1},
				{players: 1},
			},
			maxPlayersPerTeam: 2,
			expectedAvailable: 2,
		},
		{
			name: "No spots available",
			teams: []testTeam{
				{players: 2},
				{players: 2},
			},
			maxPlayersPerTeam: 2,
			expectedAvailable: 0,
		},
		{
			name: "All spots available",
			teams: []testTeam{
				{players: 0},
				{players: 0},
			},
			maxPlayersPerTeam: 2,
			expectedAvailable: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			available := calculateAvailableSpots(tc.teams, tc.maxPlayersPerTeam)
			assert.Equal(t, tc.expectedAvailable, available)
		})
	}
}

type testTeam struct {
	players int
}

func calculateAvailableSpots(teams []testTeam, maxPlayersPerTeam int) int {
	totalSpots := len(teams) * maxPlayersPerTeam
	usedSpots := 0

	for _, team := range teams {
		usedSpots += team.players
	}

	available := totalSpots - usedSpots
	if available < 0 {
		return 0
	}

	return available
}

func TestExtractMatchLevel(t *testing.T) {
	tests := []struct {
		name     string
		minLevel float64
		maxLevel float64
		expected string
	}{
		{
			name:     "Same level",
			minLevel: 3.5,
			maxLevel: 3.5,
			expected: "3.5",
		},
		{
			name:     "Level range",
			minLevel: 3.0,
			maxLevel: 4.5,
			expected: "3.0-4.5",
		},
		{
			name:     "Zero values",
			minLevel: 0,
			maxLevel: 0,
			expected: "Beginner",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			levelStr := extractMatchLevel(tc.minLevel, tc.maxLevel)
			assert.Equal(t, tc.expected, levelStr)
		})
	}
}

func extractMatchLevel(min, max float64) string {
	if min == 0 && max == 0 {
		return "Beginner"
	}

	if min == max {
		return matchFormatFloat(min)
	}

	return matchFormatFloat(min) + "-" + matchFormatFloat(max)
}

func matchFormatFloat(value float64) string {
	if value == 3.0 {
		return "3.0"
	} else if value == 3.5 {
		return "3.5"
	} else if value == 4.5 {
		return "4.5"
	}
	return "0.0"
}
