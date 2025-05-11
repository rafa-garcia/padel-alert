package testutil

import (
	"context"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

// MockRuleStorage is a mock of RuleStorage interface
type MockRuleStorage struct {
	mock.Mock
}

// CreateRule mocks the creation of a new rule in storage
func (m *MockRuleStorage) CreateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

// GetRule mocks retrieving a rule by ID from storage
func (m *MockRuleStorage) GetRule(ctx context.Context, ruleID string) (*model.Rule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rule), args.Error(1)
}

// ListRules mocks retrieving all rules for a user from storage
func (m *MockRuleStorage) ListRules(ctx context.Context, userID string) ([]*model.Rule, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.Rule), args.Error(1)
}

// UpdateRule mocks updating an existing rule in storage
func (m *MockRuleStorage) UpdateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

// DeleteRule mocks deleting a rule from storage by ID
func (m *MockRuleStorage) DeleteRule(ctx context.Context, ruleID string) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

// ScheduleRule mocks scheduling a rule for future execution
func (m *MockRuleStorage) ScheduleRule(ctx context.Context, ruleID string, nextRun time.Time) error {
	args := m.Called(ctx, ruleID, nextRun)
	return args.Error(0)
}

// GetScheduledRules mocks retrieving all scheduled rules up to a specified time
func (m *MockRuleStorage) GetScheduledRules(ctx context.Context, until time.Time) ([]string, error) {
	args := m.Called(ctx, until)
	return args.Get(0).([]string), args.Error(1)
}

// MockRedisClient implements a mock Redis client for testing
type MockRedisClient struct {
	mock.Mock
}

// SIsMember mocks checking if a member exists in a Redis set
func (m *MockRedisClient) SIsMember(ctx context.Context, key string, member string) StringCmdResult {
	args := m.Called(ctx, key, member)
	return args.Get(0).(StringCmdResult)
}

// SAdd mocks adding a member to a Redis set
func (m *MockRedisClient) SAdd(ctx context.Context, key string, member string) StringCmdResult {
	args := m.Called(ctx, key, member)
	return args.Get(0).(StringCmdResult)
}

// StringCmdResult is a mock result for Redis string commands
type StringCmdResult interface {
	Result() (interface{}, error)
	Bool() (bool, error)
	Int64() (int64, error)
}
