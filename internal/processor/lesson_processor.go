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

// LessonProcessor processes tournament/lesson rules
type LessonProcessor struct {
	client    *playtomic.Client
	ruleStore storage.RuleStorage
	redis     *storage.RedisClient
}

// NewLessonProcessor creates a new lesson processor
func NewLessonProcessor(client *playtomic.Client, ruleStore storage.RuleStorage, redis *storage.RedisClient) *LessonProcessor {
	return &LessonProcessor{
		client:    client,
		ruleStore: ruleStore,
		redis:     redis,
	}
}

// Process processes a lesson/tournament rule
func (p *LessonProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	if rule.Type != "tournament" {
		return nil, fmt.Errorf("not a tournament rule")
	}

	var allActivities []model.Activity

	// We need to make a separate request for each club ID
	for _, clubID := range rule.ClubIDs {
		// Create search parameters for each club
		params := &models.SearchLessonsParams{
			Sort:                 "start_date,created_at,ASC",
			TenantID:             clubID,
			TournamentVisibility: "PUBLIC",
			Status:               "REGISTRATION_OPEN,REGISTRATION_CLOSED,IN_PROGRESS",
			Size:                 100,
			Page:                 0,
			FromStartDate:        time.Now().Format("2006-01-02") + "T00:00:00",
		}

		lessons, err := p.client.GetLessons(ctx, params)
		if err != nil {
			logger.Error("Error fetching lessons", err, "club_id", clubID)
			continue // Skip this club and try the next one
		}

		lessonActivities, err := transformer.ExternalLessonsToActivities(lessons)
		if err != nil {
			logger.Error("Error transforming lessons", err, "club_id", clubID)
			continue
		}

		// Filters
		for _, activity := range lessonActivities {
			// Skip lessons without available spots
			if activity.AvailablePlaces <= 0 {
				continue
			}

			// Apply title filter if specified
			if rule.TitleContains != nil && *rule.TitleContains != "" {
				if !lessonContainsTitle(activity.Name, *rule.TitleContains) {
					continue
				}
			}

			// Apply date filter
			if !MatchesDateFilter(activity.StartDate, rule) {
				continue
			}

			// Check if this lesson has been seen before
			if seen, _ := p.checkSeen(ctx, rule.ID, activity.ID); seen {
				continue
			}

			allActivities = append(allActivities, activity)

			if err := p.markSeen(ctx, rule.ID, activity.ID); err != nil {
				logger.Error("Failed to mark lesson as seen", err, "rule_id", rule.ID, "lesson_id", activity.ID)
			}
		}
	}

	return allActivities, nil
}

// lessonContainsTitle checks if a lesson name contains a substring, ignoring case
func lessonContainsTitle(name, substr string) bool {
	return strings.Contains(
		strings.ToLower(name),
		strings.ToLower(substr),
	)
}

// checkSeen checks if a lesson has been seen before for a rule
func (p *LessonProcessor) checkSeen(ctx context.Context, ruleID, lessonID string) (bool, error) {
	key := fmt.Sprintf("seen:%s", ruleID)
	return p.redis.Client.SIsMember(ctx, key, lessonID).Result()
}

// markSeen marks a lesson as seen for a rule
func (p *LessonProcessor) markSeen(ctx context.Context, ruleID, lessonID string) error {
	key := fmt.Sprintf("seen:%s", ruleID)
	return p.redis.Client.SAdd(ctx, key, lessonID).Err()
}
