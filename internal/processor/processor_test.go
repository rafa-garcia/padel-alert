package processor

import (
	"context"
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMatchProcessor struct {
	mock.Mock
}

func (m *mockMatchProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	args := m.Called(ctx, rule)
	return args.Get(0).([]model.Activity), args.Error(1)
}

type mockClassProcessor struct {
	mock.Mock
}

func (m *mockClassProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	args := m.Called(ctx, rule)
	return args.Get(0).([]model.Activity), args.Error(1)
}

type mockLessonProcessor struct {
	mock.Mock
}

func (m *mockLessonProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	args := m.Called(ctx, rule)
	return args.Get(0).([]model.Activity), args.Error(1)
}

func TestProcessor_Process_SpecificType(t *testing.T) {
	mockMatch := new(mockMatchProcessor)
	mockClass := new(mockClassProcessor)
	mockLesson := new(mockLessonProcessor)

	processor := &Processor{
		matchProcessor:  mockMatch,
		classProcessor:  mockClass,
		lessonProcessor: mockLesson,
	}

	matchRule := &model.Rule{Type: "match"}
	classRule := &model.Rule{Type: "class"}
	lessonRule := &model.Rule{Type: "lesson"}

	expectedMatches := []model.Activity{{ID: "match-1", Type: "MATCH_FRIENDLY"}}
	expectedClasses := []model.Activity{{ID: "class-1", Type: "ACADEMY_CLASS"}}
	expectedLessons := []model.Activity{{ID: "lesson-1", Type: "TOURNAMENT"}}

	mockMatch.On("Process", mock.Anything, matchRule).Return(expectedMatches, nil)
	mockClass.On("Process", mock.Anything, classRule).Return(expectedClasses, nil)
	mockLesson.On("Process", mock.Anything, lessonRule).Return(expectedLessons, nil)

	activities, err := processor.Process(context.Background(), matchRule)
	assert.NoError(t, err)
	assert.Equal(t, expectedMatches, activities)

	activities, err = processor.Process(context.Background(), classRule)
	assert.NoError(t, err)
	assert.Equal(t, expectedClasses, activities)

	activities, err = processor.Process(context.Background(), lessonRule)
	assert.NoError(t, err)
	assert.Equal(t, expectedLessons, activities)

	mockMatch.AssertExpectations(t)
	mockClass.AssertExpectations(t)
	mockLesson.AssertExpectations(t)
}

func TestProcessor_Process_AllTypes(t *testing.T) {
	mockMatch := new(mockMatchProcessor)
	mockClass := new(mockClassProcessor)
	mockLesson := new(mockLessonProcessor)

	processor := &Processor{
		matchProcessor:  mockMatch,
		classProcessor:  mockClass,
		lessonProcessor: mockLesson,
	}

	allRule := &model.Rule{Type: ""} // Empty type means process all

	now := time.Now()
	matchActivity := model.Activity{ID: "match-1", Type: "MATCH_FRIENDLY", StartDate: now}
	classActivity := model.Activity{ID: "class-1", Type: "ACADEMY_CLASS", StartDate: now}
	lessonActivity := model.Activity{ID: "lesson-1", Type: "TOURNAMENT", StartDate: now}

	mockMatch.On("Process", mock.Anything, mock.MatchedBy(func(r *model.Rule) bool {
		return r.Type == "match"
	})).Return([]model.Activity{matchActivity}, nil)

	mockClass.On("Process", mock.Anything, mock.MatchedBy(func(r *model.Rule) bool {
		return r.Type == "class"
	})).Return([]model.Activity{classActivity}, nil)

	mockLesson.On("Process", mock.Anything, mock.MatchedBy(func(r *model.Rule) bool {
		return r.Type == "lesson"
	})).Return([]model.Activity{lessonActivity}, nil)

	activities, err := processor.Process(context.Background(), allRule)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(activities))
	assert.Contains(t, activities, matchActivity)
	assert.Contains(t, activities, classActivity)
	assert.Contains(t, activities, lessonActivity)

	mockMatch.AssertExpectations(t)
	mockClass.AssertExpectations(t)
	mockLesson.AssertExpectations(t)
}

func TestCommonHelperFunctions(t *testing.T) {
	t.Run("Available spots check", func(t *testing.T) {
		assert.True(t, HasAvailableSpots(2, 4))
		assert.False(t, HasAvailableSpots(4, 4))
	})

	t.Run("Date filtering", func(t *testing.T) {
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		tomorrow := now.AddDate(0, 0, 1)
		dayAfter := now.AddDate(0, 0, 2)

		rule := &model.Rule{
			StartDate: &now,
			EndDate:   &dayAfter,
		}

		assert.True(t, MatchesDateFilter(tomorrow, rule))
		assert.False(t, MatchesDateFilter(yesterday, rule))
		assert.False(t, MatchesDateFilter(dayAfter.AddDate(0, 0, 1), rule))

		noConstraintRule := &model.Rule{}
		assert.True(t, MatchesDateFilter(tomorrow, noConstraintRule))
	})
}
