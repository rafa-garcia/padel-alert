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

type MockRuleStorage struct {
	mock.Mock
}

func (m *MockRuleStorage) GetRule(ctx context.Context, ruleID string) (*model.Rule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rule), args.Error(1)
}

func (m *MockRuleStorage) ListRules(ctx context.Context, userID string) ([]*model.Rule, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Rule), args.Error(1)
}

func (m *MockRuleStorage) CreateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleStorage) UpdateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleStorage) DeleteRule(ctx context.Context, ruleID string) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockRuleStorage) ScheduleRule(ctx context.Context, ruleID string, nextRun time.Time) error {
	args := m.Called(ctx, ruleID, nextRun)
	return args.Error(0)
}

func (m *MockRuleStorage) GetScheduledRules(ctx context.Context, until time.Time) ([]string, error) {
	args := m.Called(ctx, until)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

type MockRuleProcessor struct {
	mock.Mock
}

func (m *MockRuleProcessor) processRule(ctx context.Context, ruleID string) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(3)

	taskCalled := false
	done := make(chan struct{})

	pool.Start()
	pool.Submit(func() {
		taskCalled = true
		close(done)
	})

	select {
	case <-done:
		assert.True(t, taskCalled)
	case <-time.After(1 * time.Second):
		t.Fatal("Task was not executed in time")
	}

	pool.Stop()
}

func TestScheduler_ProcessSchedule(t *testing.T) {
	mockStorage := new(MockRuleStorage)
	mockProcessor := new(MockRuleProcessor)
	cfg := &config.Config{CheckInterval: 300}

	scheduler := &Scheduler{
		config:     cfg,
		ruleStore:  mockStorage,
		processor:  mockProcessor,
		workerPool: NewWorkerPool(1),
		stopCh:     make(chan struct{}),
	}

	scheduler.workerPool.Start()
	defer scheduler.workerPool.Stop()

	now := time.Now()
	ruleIDs := []string{"rule-1", "rule-2"}

	mockStorage.On("GetScheduledRules", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return t.After(now.Add(-time.Minute)) && t.Before(now.Add(time.Minute))
	})).Return(ruleIDs, nil)

	mockProcessor.On("processRule", mock.Anything, "rule-1").Return(nil)
	mockProcessor.On("processRule", mock.Anything, "rule-2").Return(nil)

	mockStorage.On("ScheduleRule", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

	scheduler.processSchedule()

	time.Sleep(100 * time.Millisecond)

	mockStorage.AssertCalled(t, "GetScheduledRules", mock.Anything, mock.Anything)
	mockStorage.AssertNumberOfCalls(t, "ScheduleRule", len(ruleIDs))
	mockProcessor.AssertNumberOfCalls(t, "processRule", len(ruleIDs))
}

func TestScheduler_Start_Stop(t *testing.T) {
	mockStorage := new(MockRuleStorage)
	cfg := &config.Config{CheckInterval: 300}

	scheduler := NewScheduler(cfg, mockStorage)

	err := scheduler.Start()
	assert.NoError(t, err)
	assert.True(t, scheduler.running)

	scheduler.Stop()
	assert.False(t, scheduler.running)
}
