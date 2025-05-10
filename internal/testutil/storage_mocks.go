package testutil

import (
	"context"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

// MockRuleStorage mocks the rule storage interface
type MockRuleStorage struct {
	mock.Mock
}

// GetRule mocks the get rule method
func (m *MockRuleStorage) GetRule(ctx context.Context, ruleID string) (*model.Rule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rule), args.Error(1)
}

// ListRules mocks the list rules method
func (m *MockRuleStorage) ListRules(ctx context.Context, userID string) ([]*model.Rule, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Rule), args.Error(1)
}

// CreateRule mocks the create rule method
func (m *MockRuleStorage) CreateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

// UpdateRule mocks the update rule method
func (m *MockRuleStorage) UpdateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

// DeleteRule mocks the delete rule method
func (m *MockRuleStorage) DeleteRule(ctx context.Context, ruleID string) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

// ScheduleRule mocks the schedule rule method
func (m *MockRuleStorage) ScheduleRule(ctx context.Context, ruleID string, nextRun time.Time) error {
	args := m.Called(ctx, ruleID, nextRun)
	return args.Error(0)
}

// GetScheduledRules mocks the get scheduled rules method
func (m *MockRuleStorage) GetScheduledRules(ctx context.Context, until time.Time) ([]string, error) {
	args := m.Called(ctx, until)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
