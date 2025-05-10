package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmailNotifier struct {
	mock.Mock
}

func (m *MockEmailNotifier) NotifyNewActivities(ctx context.Context, user *model.User, rule *model.Rule, activities []model.Activity) error {
	args := m.Called(ctx, user, rule, activities)
	return args.Error(0)
}

type MockPlaytomicClient struct {
	mock.Mock
}

func (m *MockPlaytomicClient) SearchMatches(ctx context.Context, params interface{}) ([]model.Activity, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Activity), args.Error(1)
}

func (m *MockPlaytomicClient) SearchClasses(ctx context.Context, params interface{}) ([]model.Activity, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Activity), args.Error(1)
}

func TestRuleProcessor_ProcessRule(t *testing.T) {
	mockRuleStorage := new(MockRuleStorage)
	mockEmailNotifier := new(MockEmailNotifier)

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

	processor := &ruleProcessor{
		config:        &config.Config{},
		ruleStore:     mockRuleStorage,
		emailNotifier: mockEmailNotifier,
	}

	err := processor.processRule(context.Background(), "test-rule-id")

	assert.NoError(t, err)
	mockRuleStorage.AssertExpectations(t)
	mockEmailNotifier.AssertNotCalled(t, "NotifyNewActivities")
}

func TestRuleProcessor_InactiveRule(t *testing.T) {
	mockRuleStorage := new(MockRuleStorage)
	mockEmailNotifier := new(MockEmailNotifier)

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
	}

	err := processor.processRule(context.Background(), "test-rule-id")

	assert.NoError(t, err)
	mockRuleStorage.AssertCalled(t, "GetRule", mock.Anything, "test-rule-id")
	mockRuleStorage.AssertNotCalled(t, "UpdateRule", mock.Anything, mock.Anything)
	mockEmailNotifier.AssertNotCalled(t, "NotifyNewActivities")
}

func TestRuleProcessor_RuleNotFound(t *testing.T) {
	mockRuleStorage := new(MockRuleStorage)
	mockEmailNotifier := new(MockEmailNotifier)

	mockRuleStorage.On("GetRule", mock.Anything, "non-existent-rule").Return(nil, nil)

	processor := &ruleProcessor{
		config:        &config.Config{},
		ruleStore:     mockRuleStorage,
		emailNotifier: mockEmailNotifier,
	}

	err := processor.processRule(context.Background(), "non-existent-rule")

	assert.NoError(t, err)
	mockRuleStorage.AssertCalled(t, "GetRule", mock.Anything, "non-existent-rule")
	mockRuleStorage.AssertNotCalled(t, "UpdateRule", mock.Anything, mock.Anything)
	mockEmailNotifier.AssertNotCalled(t, "NotifyNewActivities")
}
