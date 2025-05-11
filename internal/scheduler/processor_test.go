package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRuleProcessor_ProcessRule_NoMatches(t *testing.T) {
	mockRuleStorage := new(testutil.MockRuleStorage)
	mockEmailNotifier := new(testutil.MockEmailNotifier)
	mockProcessor := new(testutil.MockProcessor)

	now := time.Now()
	rule := &model.Rule{
		ID:        "test-rule-id",
		UserID:    "test-user-id",
		Email:     "test@example.com",
		Type:      "match",
		Name:      "Test Rule",
		ClubIDs:   []string{"club-1"},
		Active:    true,
		CreatedAt: now,
	}

	mockRuleStorage.On("GetRule", mock.Anything, "test-rule-id").Return(rule, nil)
	mockRuleStorage.On("UpdateRule", mock.Anything, mock.MatchedBy(func(r *model.Rule) bool {
		return r.ID == "test-rule-id" && !r.LastChecked.IsZero()
	})).Return(nil)

	mockProcessor.On("Process", mock.Anything, rule).Return([]model.Activity{}, nil)

	processor := &ruleProcessor{
		config:        &config.Config{},
		ruleStore:     mockRuleStorage,
		emailNotifier: mockEmailNotifier,
		processor:     mockProcessor,
	}

	err := processor.processRule(context.Background(), "test-rule-id")

	assert.NoError(t, err)
	mockRuleStorage.AssertExpectations(t)
	mockProcessor.AssertExpectations(t)
	mockEmailNotifier.AssertNotCalled(t, "NotifyNewActivities")
}

func TestRuleProcessor_ProcessRule_WithMatches(t *testing.T) {
	mockRuleStorage := new(testutil.MockRuleStorage)
	mockEmailNotifier := new(testutil.MockEmailNotifier)
	mockProcessor := new(testutil.MockProcessor)

	now := time.Now()
	rule := &model.Rule{
		ID:        "test-rule-id",
		UserID:    "test-user-id",
		Email:     "test@example.com",
		Type:      "match",
		Name:      "Test Rule",
		ClubIDs:   []string{"club-1"},
		Active:    true,
		CreatedAt: now,
	}

	activities := []model.Activity{
		{
			ID:   "activity-1",
			Name: "Test Match",
			Club: model.Club{
				ID:   "club-1",
				Name: "Test Club",
			},
		},
	}

	mockRuleStorage.On("GetRule", mock.Anything, "test-rule-id").Return(rule, nil)
	mockRuleStorage.On("UpdateRule", mock.Anything, mock.MatchedBy(func(r *model.Rule) bool {
		return r.ID == "test-rule-id" && !r.LastChecked.IsZero() && !r.LastNotification.IsZero()
	})).Return(nil)

	mockProcessor.On("Process", mock.Anything, rule).Return(activities, nil)

	mockEmailNotifier.On("NotifyNewActivities", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.ID == "test-user-id" && u.Email == "test@example.com"
	}), rule, activities).Return(nil)

	processor := &ruleProcessor{
		config:        &config.Config{},
		ruleStore:     mockRuleStorage,
		emailNotifier: mockEmailNotifier,
		processor:     mockProcessor,
	}

	err := processor.processRule(context.Background(), "test-rule-id")

	assert.NoError(t, err)
	mockRuleStorage.AssertExpectations(t)
	mockProcessor.AssertExpectations(t)
	mockEmailNotifier.AssertExpectations(t)
}

func TestRuleProcessor_InactiveRule(t *testing.T) {
	mockRuleStorage := new(testutil.MockRuleStorage)
	mockEmailNotifier := new(testutil.MockEmailNotifier)
	mockProcessor := new(testutil.MockProcessor)

	rule := &model.Rule{
		ID:      "test-rule-id",
		UserID:  "test-user-id",
		Email:   "test@example.com",
		Type:    "match",
		Name:    "Test Rule",
		ClubIDs: []string{"club-1"},
		Active:  false, // Inactive rule
	}

	mockRuleStorage.On("GetRule", mock.Anything, "test-rule-id").Return(rule, nil)

	processor := &ruleProcessor{
		config:        &config.Config{},
		ruleStore:     mockRuleStorage,
		emailNotifier: mockEmailNotifier,
		processor:     mockProcessor,
	}

	err := processor.processRule(context.Background(), "test-rule-id")

	assert.NoError(t, err)
	mockRuleStorage.AssertExpectations(t)
	mockProcessor.AssertNotCalled(t, "Process")
	mockEmailNotifier.AssertNotCalled(t, "NotifyNewActivities")
}
