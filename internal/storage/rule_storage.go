package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/redis/go-redis/v9"
)

// RuleStorage defines operations for rule persistence
type RuleStorage interface {
	GetRule(ctx context.Context, ruleID string) (*model.Rule, error)
	ListRules(ctx context.Context, userID string) ([]*model.Rule, error)
	CreateRule(ctx context.Context, rule *model.Rule) error
	UpdateRule(ctx context.Context, rule *model.Rule) error
	DeleteRule(ctx context.Context, ruleID string) error
	ScheduleRule(ctx context.Context, ruleID string, nextRun time.Time) error
	GetScheduledRules(ctx context.Context, until time.Time) ([]string, error)
}

// RedisRuleStorage implements RuleStorage using Redis
type RedisRuleStorage struct {
	redis *RedisClient
}

// NewRedisRuleStorage creates a new Redis rule storage
func NewRedisRuleStorage(redis *RedisClient) *RedisRuleStorage {
	return &RedisRuleStorage{redis: redis}
}

// GetRule gets a rule by ID
func (s *RedisRuleStorage) GetRule(ctx context.Context, ruleID string) (*model.Rule, error) {
	key := fmt.Sprintf("rule:%s", ruleID)
	data, err := s.redis.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get rule: %w", err)
	}

	var rule model.Rule
	if err := json.Unmarshal([]byte(data), &rule); err != nil {
		return nil, fmt.Errorf("unmarshal rule: %w", err)
	}

	return &rule, nil
}

// ListRules lists rules for a user
func (s *RedisRuleStorage) ListRules(ctx context.Context, userID string) ([]*model.Rule, error) {
	key := fmt.Sprintf("rules:user:%s", userID)
	ruleIDs, err := s.redis.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("list user rules: %w", err)
	}

	rules := make([]*model.Rule, 0, len(ruleIDs))
	for _, id := range ruleIDs {
		rule, err := s.GetRule(ctx, id)
		if err != nil {
			continue // Skip rules that can't be loaded
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// CreateRule stores a new rule
func (s *RedisRuleStorage) CreateRule(ctx context.Context, rule *model.Rule) error {
	// Set timestamps
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	// Store rule
	key := fmt.Sprintf("rule:%s", rule.ID)
	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("marshal rule: %w", err)
	}

	// Add to user's rule set
	userKey := fmt.Sprintf("rules:user:%s", rule.UserID)

	// Execute in pipeline
	pipe := s.redis.Client.Pipeline()
	pipe.Set(ctx, key, data, 0)
	pipe.SAdd(ctx, userKey, rule.ID)
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("create rule: %w", err)
	}

	return nil
}

// UpdateRule updates an existing rule
func (s *RedisRuleStorage) UpdateRule(ctx context.Context, rule *model.Rule) error {
	// Check if rule exists
	key := fmt.Sprintf("rule:%s", rule.ID)
	exists, err := s.redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("check rule existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	// Update timestamp
	rule.UpdatedAt = time.Now()

	// Serialize and store
	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("marshal rule: %w", err)
	}

	if err := s.redis.Client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("update rule: %w", err)
	}

	return nil
}

// DeleteRule removes a rule
func (s *RedisRuleStorage) DeleteRule(ctx context.Context, ruleID string) error {
	// Get rule to check owner
	rule, err := s.GetRule(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("get rule: %w", err)
	}

	key := fmt.Sprintf("rule:%s", ruleID)
	userKey := fmt.Sprintf("rules:user:%s", rule.UserID)
	scheduleKey := "rules:schedule"
	seenKey := fmt.Sprintf("seen:%s", ruleID)

	pipe := s.redis.Client.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, userKey, ruleID)
	pipe.ZRem(ctx, scheduleKey, ruleID)
	pipe.Del(ctx, seenKey)
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}

	return nil
}

// ScheduleRule schedules a rule for execution
func (s *RedisRuleStorage) ScheduleRule(ctx context.Context, ruleID string, nextRun time.Time) error {
	key := "rules:schedule"
	score := float64(nextRun.Unix())

	if err := s.redis.Client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: ruleID,
	}).Err(); err != nil {
		return fmt.Errorf("schedule rule: %w", err)
	}

	return nil
}

// GetScheduledRules gets rules scheduled until a specific time
func (s *RedisRuleStorage) GetScheduledRules(ctx context.Context, until time.Time) ([]string, error) {
	key := "rules:schedule"
	max := float64(until.Unix())

	rules, err := s.redis.Client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%f", max),
		Count: 100, // Limit batch size
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("get scheduled rules: %w", err)
	}

	return rules, nil
}
