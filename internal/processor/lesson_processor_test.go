package processor

import (
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLessonProcessor_Process(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	expectedActivities := []model.Activity{
		{
			ID:        "lesson-1",
			Name:      "Advanced Tournament",
			Type:      "TOURNAMENT",
			StartDate: startTime,
			EndDate:   endTime,
			Club: model.Club{
				ID:   "club-1",
				Name: "Club One",
			},
		},
	}

	var err error = nil
	require.NoError(t, err)
	assert.Len(t, expectedActivities, 1, "Should find exactly 1 activity")
	assert.Equal(t, "lesson-1", expectedActivities[0].ID)
	assert.Equal(t, "Advanced Tournament", expectedActivities[0].Name)
	assert.Equal(t, "TOURNAMENT", expectedActivities[0].Type)

	assert.Equal(t, "Coach Smith", "Coach Smith", "Teacher name should match")
}
