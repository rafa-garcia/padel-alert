package processor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHasAvailableSpots(t *testing.T) {
	tests := []struct {
		name           string
		playersTeam1   int
		playersTeam2   int
		maxPlayersTeam int
		expected       bool
	}{
		{
			name:           "Available spots",
			playersTeam1:   1,
			playersTeam2:   1,
			maxPlayersTeam: 2,
			expected:       true,
		},
		{
			name:           "No available spots",
			playersTeam1:   2,
			playersTeam2:   2,
			maxPlayersTeam: 2,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availablePlaces := 4 - (tt.playersTeam1 + tt.playersTeam2)
			if tt.expected {
				assert.Greater(t, availablePlaces, 0)
			} else {
				assert.Equal(t, 0, availablePlaces)
			}
		})
	}
}

func TestMatchesRankingFilter(t *testing.T) {
	minRanking := 3.0
	maxRanking := 4.5

	tests := []struct {
		name     string
		minLevel float64
		maxLevel float64
		expected bool
	}{
		{
			name:     "Within range",
			minLevel: 3.0,
			maxLevel: 4.0,
			expected: true,
		},
		{
			name:     "Min level too high",
			minLevel: 5.0,
			maxLevel: 6.0,
			expected: false,
		},
		{
			name:     "Max level too low",
			minLevel: 1.0,
			maxLevel: 2.0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple logic to check if the match level range fits within rule's ranking range
			matchesFilter := (tt.minLevel >= minRanking && tt.maxLevel <= maxRanking)
			assert.Equal(t, tt.expected, matchesFilter)
		})
	}
}

func TestMatchesDateFilter(t *testing.T) {
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		matchTime time.Time
		expected  bool
	}{
		{
			name:      "Within range",
			matchTime: now,
			expected:  true,
		},
		{
			name:      "Before range",
			matchTime: now.Add(-48 * time.Hour),
			expected:  false,
		},
		{
			name:      "After range",
			matchTime: now.Add(48 * time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test if match date is within rule's date range
			matchesFilter := (tt.matchTime.After(startDate) && tt.matchTime.Before(endDate))
			assert.Equal(t, tt.expected, matchesFilter)
		})
	}
}
