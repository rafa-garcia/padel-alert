package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	playtomic "github.com/rafa-garcia/go-playtomic-api/client"
	"github.com/rafa-garcia/go-playtomic-api/models"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/storage"
	"github.com/rafa-garcia/padel-alert/internal/transformer"
)

// ClassProcessor processes class rules
type ClassProcessor struct {
	client    *playtomic.Client
	ruleStore storage.RuleStorage
	redis     *storage.RedisClient
}

// NewClassProcessor creates a new class processor
func NewClassProcessor(client *playtomic.Client, ruleStore storage.RuleStorage, redis *storage.RedisClient) *ClassProcessor {
	return &ClassProcessor{
		client:    client,
		ruleStore: ruleStore,
		redis:     redis,
	}
}

// Process processes a class rule
func (p *ClassProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	if rule.Type != "class" {
		return nil, fmt.Errorf("not a class rule")
	}

	// Create search parameters
	params := &models.SearchClassesParams{
		Sort:             "start_date,created_at,ASC",
		Status:           "PENDING,IN_PROGRESS", // Only active classes
		TenantIDs:        rule.ClubIDs,
		IncludeSummary:   true,
		Size:             100,
		Page:             0,
		CourseVisibility: "PUBLIC",
		FromStartDate:    time.Now().Format("2006-01-02") + "T00:00:00",
	}

	classes, err := p.client.GetClasses(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch classes: %w", err)
	}

	classActivities, err := transformer.ExternalClassesToActivities(classes)
	if err != nil {
		return nil, fmt.Errorf("transform classes: %w", err)
	}

	var activities []model.Activity

	// Filters
	for _, activity := range classActivities {
		// Skip classes without available spots
		if activity.AvailablePlaces <= 0 {
			continue
		}

		// Apply title filter if specified
		if rule.TitleContains != nil && *rule.TitleContains != "" {
			if !containsIgnoreCase(activity.Name, *rule.TitleContains) {
				continue
			}
		}

		// Apply date filter
		if !MatchesDateFilter(activity.StartDate, rule) {
			continue
		}

		// Check if this class has been seen before
		if seen, _ := p.checkSeen(ctx, rule.ID, activity.ID); seen {
			continue
		}

		activities = append(activities, activity)

		if err := p.markSeen(ctx, rule.ID, activity.ID); err != nil {
			logger.Error("Failed to mark class as seen", err, "rule_id", rule.ID, "class_id", activity.ID)
		}
	}

	return activities, nil
}

// containsIgnoreCase checks if a string contains a substring, ignoring case
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

// checkSeen checks if a class has been seen before for a rule
func (p *ClassProcessor) checkSeen(ctx context.Context, ruleID string, classID string) (bool, error) {
	key := fmt.Sprintf("seen:%s", ruleID)
	return p.redis.Client.SIsMember(ctx, key, classID).Result()
}

// markSeen marks a class as seen for a rule
func (p *ClassProcessor) markSeen(ctx context.Context, ruleID string, classID string) error {
	key := fmt.Sprintf("seen:%s", ruleID)
	return p.redis.Client.SAdd(ctx, key, classID).Err()
}
