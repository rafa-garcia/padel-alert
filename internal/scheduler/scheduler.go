package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/storage"
)

// Scheduler manages periodic rule processing
type Scheduler struct {
	config     *config.Config
	ruleStore  storage.RuleStorage
	workerPool *WorkerPool
	stopCh     chan struct{}
	wg         sync.WaitGroup
	running    bool
	mu         sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.Config, ruleStore storage.RuleStorage) *Scheduler {
	return &Scheduler{
		config:     cfg,
		ruleStore:  ruleStore,
		workerPool: NewWorkerPool(10), // 10 workers
		stopCh:     make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.workerPool.Start()
	s.running = true

	s.wg.Add(1)
	go s.run()

	logger.Info("Scheduler started", "interval", s.config.CheckInterval)
	return nil
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopCh)
	s.wg.Wait()
	s.workerPool.Stop()
	s.running = false

	logger.Info("Scheduler stopped")
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processSchedule()
		case <-s.stopCh:
			return
		}
	}
}

// processSchedule checks and processes scheduled rules
func (s *Scheduler) processSchedule() {
	ctx := context.Background()
	now := time.Now()

	// Get rules scheduled until now
	rules, err := s.ruleStore.GetScheduledRules(ctx, now)
	if err != nil {
		logger.Error("Failed to get scheduled rules", err)
		return
	}

	for _, ruleID := range rules {
		ruleID := ruleID // Create new variable for closure

		// Submit to worker pool
		s.workerPool.Submit(func() {
			// Process rule here (will be implemented later)
			// For now, just reschedule it
			next := time.Now().Add(time.Duration(s.config.CheckInterval) * time.Second)
			if err := s.ruleStore.ScheduleRule(ctx, ruleID, next); err != nil {
				logger.Error("Failed to reschedule rule", err, "ruleID", ruleID)
			}
		})
	}
}

// WorkerPool handles concurrent task processing
type WorkerPool struct {
	numWorkers int
	tasks      chan func()
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		tasks:      make(chan func(), 100),
		stopCh:     make(chan struct{}),
	}
}

// Start begins the worker pool
func (p *WorkerPool) Start() {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop gracefully stops the worker pool
func (p *WorkerPool) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}

// Submit adds a task to the pool
func (p *WorkerPool) Submit(task func()) {
	select {
	case p.tasks <- task:
	case <-p.stopCh:
	}
}

// worker processes tasks
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case task := <-p.tasks:
			// Execute the task, catch panics
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("Task panicked", fmt.Errorf("panic: %v", r))
					}
				}()
				task()
			}()
		case <-p.stopCh:
			return
		}
	}
}
