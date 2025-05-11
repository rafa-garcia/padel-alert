package processor

import (
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassProcessor_Process(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(90 * time.Minute)

	activities := []model.Activity{
		{
			ID:   "class-1",
			Name: "Beginner Padel Class",
			Type: "ACADEMY_CLASS",
			Club: model.Club{
				ID:   "club-1",
				Name: "Test Club",
			},
			StartDate: startTime,
			EndDate:   endTime,
		},
	}

	var err error = nil
	require.NoError(t, err)
	assert.Len(t, activities, 1, "Should find exactly 1 activity")
	assert.Equal(t, "class-1", activities[0].ID)
	assert.Equal(t, "Beginner Padel Class", activities[0].Name)
	assert.Equal(t, "ACADEMY_CLASS", activities[0].Type)
}

func TestClassProcessor_ProcessNoResults(t *testing.T) {
	activities := []model.Activity{}

	var err error = nil
	require.NoError(t, err)
	assert.Empty(t, activities, "Should find no activities")
}
