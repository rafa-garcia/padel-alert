package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// testRuleProcessor is a simple test implementation of RuleProcessor
// used only for testing the scheduler
type testRuleProcessor struct {
	processCalled bool
	processError  error
	processedIDs  []string
	mu            sync.Mutex
}

func newTestRuleProcessor() *testRuleProcessor {
	return &testRuleProcessor{
		processCalled: false,
		processedIDs:  make([]string, 0),
	}
}

func (p *testRuleProcessor) processRule(ctx context.Context, ruleID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.processCalled = true
	p.processedIDs = append(p.processedIDs, ruleID)
	return p.processError
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

func TestScheduler_Start_Stop(t *testing.T) {
	mockStorage := new(testutil.MockRuleStorage)
	cfg := &config.Config{CheckInterval: 300}

	scheduler := NewScheduler(cfg, mockStorage)

	err := scheduler.Start()
	assert.NoError(t, err)
	assert.True(t, scheduler.running)

	err = scheduler.Start()
	assert.Error(t, err)

	scheduler.Stop()
	assert.False(t, scheduler.running)
}

func TestScheduler_ProcessSchedule(t *testing.T) {
	mockStorage := new(testutil.MockRuleStorage)
	testProcessor := newTestRuleProcessor()
	cfg := &config.Config{CheckInterval: 300}

	scheduler := &Scheduler{
		config:     cfg,
		ruleStore:  mockStorage,
		processor:  testProcessor,
		workerPool: NewWorkerPool(1),
		stopCh:     make(chan struct{}),
	}

	scheduler.workerPool.Start()
	defer scheduler.workerPool.Stop()
	ruleIDs := []string{"rule-1", "rule-2"}

	mockStorage.On("GetScheduledRules", mock.Anything, mock.Anything).Return(ruleIDs, nil)

	mockStorage.On("ScheduleRule", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	scheduler.processSchedule()

	time.Sleep(100 * time.Millisecond)

	testProcessor.mu.Lock()
	processedCount := len(testProcessor.processedIDs)
	processedIDs := make([]string, len(testProcessor.processedIDs))
	copy(processedIDs, testProcessor.processedIDs)
	testProcessor.mu.Unlock()

	assert.Equal(t, 2, processedCount, "Should have processed 2 rules")
	assert.Contains(t, processedIDs, "rule-1", "Should have processed rule-1")
	assert.Contains(t, processedIDs, "rule-2", "Should have processed rule-2")

	mockStorage.AssertExpectations(t)
}
