package processor

import (
	"context"
	"sync"
	"time"

	playtomic "github.com/rafa-garcia/go-playtomic-api/client"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/storage"
)

// ProcessorInterface defines the interface for activity processing
type ProcessorInterface interface {
	Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error)
}

// HasAvailableSpots checks if an activity has available spots
func HasAvailableSpots(currentPlayers, maxPlayers int) bool {
	return currentPlayers < maxPlayers
}

// MatchesDateFilter checks if a date falls within the rule's date constraints
func MatchesDateFilter(date time.Time, rule *model.Rule) bool {
	if rule.StartDate != nil && date.Before(*rule.StartDate) {
		return false
	}

	if rule.EndDate != nil && date.After(*rule.EndDate) {
		return false
	}

	return true
}

// Processor processes rules for all activity types
type Processor struct {
	matchProcessor  ProcessorInterface
	classProcessor  ProcessorInterface
	lessonProcessor ProcessorInterface
}

// NewProcessor creates a new processor that handles all activity types
func NewProcessor(client *playtomic.Client, ruleStore storage.RuleStorage, redis *storage.RedisClient) *Processor {
	return &Processor{
		matchProcessor:  NewMatchProcessor(client, ruleStore, redis),
		classProcessor:  NewClassProcessor(client, ruleStore, redis),
		lessonProcessor: NewLessonProcessor(client, ruleStore, redis),
	}
}

// Process processes a rule for all activity types concurrently
func (p *Processor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	// If rule type is specific, delegate to the appropriate processor
	switch rule.Type {
	case "match":
		return p.matchProcessor.Process(ctx, rule)
	case "class":
		return p.classProcessor.Process(ctx, rule)
	case "tournament":
		return p.lessonProcessor.Process(ctx, rule)
	}

	// For empty or "all" type, process all types concurrently
	var allActivities []model.Activity
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	matchRule := *rule
	matchRule.Type = "match"

	classRule := *rule
	classRule.Type = "class"

	tournamentRule := *rule
	tournamentRule.Type = "tournament"

	// Process matches
	wg.Add(1)
	go func() {
		defer wg.Done()
		activities, err := p.matchProcessor.Process(ctx, &matchRule)
		if err != nil {
			logger.Error("Failed to process matches", err, "rule_id", rule.ID)
			errCh <- err
			return
		}

		mu.Lock()
		allActivities = append(allActivities, activities...)
		mu.Unlock()
	}()

	// Process classes
	wg.Add(1)
	go func() {
		defer wg.Done()
		activities, err := p.classProcessor.Process(ctx, &classRule)
		if err != nil {
			logger.Error("Failed to process classes", err, "rule_id", rule.ID)
			errCh <- err
			return
		}

		mu.Lock()
		allActivities = append(allActivities, activities...)
		mu.Unlock()
	}()

	// Process tournaments/lessons
	wg.Add(1)
	go func() {
		defer wg.Done()
		activities, err := p.lessonProcessor.Process(ctx, &tournamentRule)
		if err != nil {
			logger.Error("Failed to process tournaments", err, "rule_id", rule.ID)
			errCh <- err
			return
		}

		mu.Lock()
		allActivities = append(allActivities, activities...)
		mu.Unlock()
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		if len(allActivities) > 0 {
			return allActivities, nil
		}
		return nil, err
	default:
		return allActivities, nil
	}
}
