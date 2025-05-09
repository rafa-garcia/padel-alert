package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestActivity(t *testing.T) {
	now := time.Now()

	activity := Activity{
		ID:              "test-activity",
		Type:            "MATCH_COMPETITIVE",
		PlaytomicType:   "COMPETITIVE",
		Name:            "Test Match",
		StartDate:       now,
		EndDate:         now.Add(time.Hour),
		Duration:        60,
		AvailablePlaces: 2,
		RegisteredPlayers: []Player{
			{
				ID:    "player-1",
				Name:  "Player 1",
				Level: 3.5,
				Link:  "https://example.com/player-1",
			},
			{
				ID:    "player-2",
				Name:  "Player 2",
				Level: 4.0,
				Link:  "https://example.com/player-2",
			},
		},
		Club: Club{
			ID:   "club-1",
			Name: "Test Club",
			Address: Address{
				Street:     "123 Main St",
				PostalCode: "12345",
				City:       "Test City",
				Country:    "Test Country",
			},
			Link: "https://example.com/club-1",
		},
	}

	assert.Equal(t, "test-activity", activity.ID)
	assert.Equal(t, "MATCH_COMPETITIVE", activity.Type)
	assert.Equal(t, "COMPETITIVE", activity.PlaytomicType)
	assert.Equal(t, "Test Match", activity.Name)
	assert.Equal(t, now, activity.StartDate)
	assert.Equal(t, now.Add(time.Hour), activity.EndDate)
	assert.Equal(t, 60, activity.Duration)
	assert.Equal(t, 2, activity.AvailablePlaces)
	assert.Equal(t, 2, len(activity.RegisteredPlayers))

	assert.Equal(t, "club-1", activity.Club.ID)
	assert.Equal(t, "Test Club", activity.Club.Name)
	assert.Equal(t, "Test City", activity.Club.Address.City)

	assert.Equal(t, "player-1", activity.RegisteredPlayers[0].ID)
	assert.Equal(t, "Player 1", activity.RegisteredPlayers[0].Name)
	assert.Equal(t, 3.5, activity.RegisteredPlayers[0].Level)
}

func TestPlayer(t *testing.T) {
	teamA := "A"

	player := Player{
		ID:    "player-1",
		Name:  "Player 1",
		Level: 3.5,
		Team:  &teamA,
		Link:  "https://example.com/player-1",
	}

	assert.Equal(t, "player-1", player.ID)
	assert.Equal(t, "Player 1", player.Name)
	assert.Equal(t, 3.5, player.Level)
	assert.Equal(t, "A", *player.Team)
	assert.Equal(t, "https://example.com/player-1", player.Link)
}

func TestClub(t *testing.T) {
	club := Club{
		ID:   "club-1",
		Name: "Test Club",
		Address: Address{
			Street:     "123 Main St",
			PostalCode: "12345",
			City:       "Test City",
			Country:    "Test Country",
		},
	}

	assert.Equal(t, "club-1", club.ID)
	assert.Equal(t, "Test Club", club.Name)
	assert.Equal(t, "123 Main St", club.Address.Street)
	assert.Equal(t, "12345", club.Address.PostalCode)
	assert.Equal(t, "Test City", club.Address.City)
	assert.Equal(t, "Test Country", club.Address.Country)
}
