package scheduler

import (
	"context"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/notification"
	"github.com/rafa-garcia/padel-alert/internal/processor"
	"github.com/rafa-garcia/padel-alert/internal/storage"

	playtomic "github.com/rafa-garcia/go-playtomic-api/client"
)

// EmailNotifier defines an interface for email notification
type EmailNotifier interface {
	NotifyNewActivities(ctx context.Context, user *model.User, rule *model.Rule, activities []model.Activity) error
}

// RuleProcessor defines the interface for processing rules
type RuleProcessor interface {
	processRule(ctx context.Context, ruleID string) error
}

// PlaytomicClient interface for Playtomic API operations
type PlaytomicClient interface {
	SearchMatches(ctx context.Context, params interface{}) ([]model.Activity, error)
	SearchClasses(ctx context.Context, params interface{}) ([]model.Activity, error)
}

// MatchProcessor interface for processing match rules
type MatchProcessor interface {
	Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error)
}

// ruleProcessor processes rules and sends notifications
type ruleProcessor struct {
	config         *config.Config
	ruleStore      storage.RuleStorage
	emailNotifier  EmailNotifier
	matchProcessor MatchProcessor
}

// newRuleProcessor creates a new rule processor
func newRuleProcessor(cfg *config.Config, ruleStore storage.RuleStorage) *ruleProcessor {
	redisClient, err := storage.NewRedisClient(cfg)
	if err != nil {
		logger.Error("Failed to create Redis client", err)
		redisClient = nil
	}

	playtomicClient := playtomic.NewClient(
		playtomic.WithTimeout(60*time.Second),
		playtomic.WithRetries(3),
	)

	return &ruleProcessor{
		config:         cfg,
		ruleStore:      ruleStore,
		emailNotifier:  notification.NewEmailNotifier(cfg),
		matchProcessor: processor.NewMatchProcessor(playtomicClient, ruleStore, redisClient),
	}
}

// processRule processes a rule and sends notifications if needed
func (p *ruleProcessor) processRule(ctx context.Context, ruleID string) error {
	rule, err := p.ruleStore.GetRule(ctx, ruleID)
	if err != nil {
		logger.Error("Failed to get rule", err, "rule_id", ruleID)
		return err
	}

	if rule == nil {
		logger.Warn("Rule not found", "rule_id", ruleID)
		return nil
	}

	if !rule.Active {
		logger.Debug("Skipping inactive rule", "rule_id", ruleID, "name", rule.Name)
		return nil
	}

	logger.Debug("Processing rule", "rule_id", ruleID, "name", rule.Name, "type", rule.Type)

	rule.LastChecked = time.Now()

	var activities []model.Activity

	if rule.Type == "match" {
		matchActivities, err := p.matchProcessor.Process(ctx, rule)
		if err != nil {
			logger.Error("Failed to process match rule", err, "rule_id", ruleID)
		} else {
			activities = matchActivities
		}
	} else {
		logger.Warn("Unsupported rule type", "rule_id", ruleID, "type", rule.Type)
	}

	if len(activities) > 0 {
		user := &model.User{
			ID:    rule.UserID,
			Email: rule.Email,
		}

		logger.Info("Sending notification", "rule_id", ruleID, "activities", len(activities))
		err = p.emailNotifier.NotifyNewActivities(ctx, user, rule, activities)
		if err != nil {
			logger.Error("Failed to send notification", err, "rule_id", ruleID)
		} else {
			rule.LastNotification = time.Now()
			logger.Info("Notification sent successfully", "rule_id", ruleID)
		}
	} else {
		logger.Info("No activities found for rule", "rule_id", ruleID)
	}

	err = p.ruleStore.UpdateRule(ctx, rule)
	if err != nil {
		logger.Error("Failed to update rule", err, "rule_id", ruleID)
	}

	return nil
}
