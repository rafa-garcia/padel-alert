package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *RedisClient) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mini.Addr(),
	})

	redisClient := &RedisClient{
		Client: client,
	}

	return mini, redisClient
}

func TestRedisRuleStorage_CreateRule(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	ruleStorage := NewRedisRuleStorage(redisClient)
	ctx := context.Background()

	rule := &model.Rule{
		ID:     "test-rule-123",
		UserID: "test-user-456",
		Type:   "match",
		Name:   "Test Rule",
		ClubIDs: []string{
			"club-1",
			"club-2",
		},
	}

	err := ruleStorage.CreateRule(ctx, rule)
	assert.NoError(t, err)

	exists := mini.Exists("rule:" + rule.ID)
	assert.True(t, exists)

	userRuleKey := "rules:user:" + rule.UserID
	members, err := mini.SMembers(userRuleKey)
	assert.NoError(t, err)
	assert.Contains(t, members, rule.ID)
}

func TestRedisRuleStorage_GetRule(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	ruleStorage := NewRedisRuleStorage(redisClient)
	ctx := context.Background()

	now := time.Now().UTC()
	minRanking := 3.0
	maxRanking := 4.5
	rule := &model.Rule{
		ID:         "test-rule-123",
		UserID:     "test-user-456",
		Type:       "match",
		Name:       "Test Rule",
		ClubIDs:    []string{"club-1", "club-2"},
		MinRanking: &minRanking,
		MaxRanking: &maxRanking,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := ruleStorage.CreateRule(ctx, rule)
	assert.NoError(t, err)

	fetchedRule, err := ruleStorage.GetRule(ctx, rule.ID)
	assert.NoError(t, err)

	assert.Equal(t, rule.ID, fetchedRule.ID)
	assert.Equal(t, rule.UserID, fetchedRule.UserID)
	assert.Equal(t, rule.Type, fetchedRule.Type)
	assert.Equal(t, rule.Name, fetchedRule.Name)
	assert.Equal(t, rule.ClubIDs, fetchedRule.ClubIDs)
	assert.Equal(t, *rule.MinRanking, *fetchedRule.MinRanking)
	assert.Equal(t, *rule.MaxRanking, *fetchedRule.MaxRanking)
}

func TestRedisRuleStorage_ListRules(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	ruleStorage := NewRedisRuleStorage(redisClient)
	ctx := context.Background()
	userID := "test-user-456"

	rule1 := &model.Rule{
		ID:      "test-rule-1",
		UserID:  userID,
		Type:    "match",
		Name:    "Test Rule 1",
		ClubIDs: []string{"club-1"},
	}
	rule2 := &model.Rule{
		ID:      "test-rule-2",
		UserID:  userID,
		Type:    "class",
		Name:    "Test Rule 2",
		ClubIDs: []string{"club-2"},
	}

	err := ruleStorage.CreateRule(ctx, rule1)
	assert.NoError(t, err)
	err = ruleStorage.CreateRule(ctx, rule2)
	assert.NoError(t, err)

	rules, err := ruleStorage.ListRules(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, rules, 2)

	ruleIDs := []string{}
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}
	assert.Contains(t, ruleIDs, rule1.ID)
	assert.Contains(t, ruleIDs, rule2.ID)
}

func TestRedisRuleStorage_UpdateRule(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	ruleStorage := NewRedisRuleStorage(redisClient)
	ctx := context.Background()

	rule := &model.Rule{
		ID:      "test-rule-123",
		UserID:  "test-user-456",
		Type:    "match",
		Name:    "Test Rule",
		ClubIDs: []string{"club-1"},
	}

	err := ruleStorage.CreateRule(ctx, rule)
	assert.NoError(t, err)

	rule.Name = "Updated Rule Name"
	rule.ClubIDs = []string{"club-1", "club-3"}
	err = ruleStorage.UpdateRule(ctx, rule)
	assert.NoError(t, err)

	updatedRule, err := ruleStorage.GetRule(ctx, rule.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Rule Name", updatedRule.Name)
	assert.Equal(t, []string{"club-1", "club-3"}, updatedRule.ClubIDs)
}

func TestRedisRuleStorage_DeleteRule(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	ruleStorage := NewRedisRuleStorage(redisClient)
	ctx := context.Background()
	userID := "test-user-456"
	ruleID := "test-rule-123"

	rule := &model.Rule{
		ID:      ruleID,
		UserID:  userID,
		Type:    "match",
		Name:    "Test Rule",
		ClubIDs: []string{"club-1"},
	}

	err := ruleStorage.CreateRule(ctx, rule)
	assert.NoError(t, err)

	seenKey := "seen:" + ruleID
	scheduleKey := "rules:schedule"
	err = mini.Set(seenKey, "some-data")
	assert.NoError(t, err)
	_, err = mini.ZAdd(scheduleKey, 1.0, ruleID)
	assert.NoError(t, err)

	err = ruleStorage.DeleteRule(ctx, ruleID)
	assert.NoError(t, err)

	exists := mini.Exists("rule:" + ruleID)
	assert.False(t, exists)

	userRuleKey := "rules:user:" + userID
	members, _ := mini.SMembers(userRuleKey)
	assert.NotContains(t, members, ruleID)

	exists = mini.Exists(seenKey)
	assert.False(t, exists)

	scheduleMembers, _ := mini.ZMembers(scheduleKey)
	assert.NotContains(t, scheduleMembers, ruleID)
}

func TestRedisRuleStorage_ScheduleRule(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	ruleStorage := NewRedisRuleStorage(redisClient)
	ctx := context.Background()
	ruleID := "test-rule-123"

	now := time.Now().UTC()
	err := ruleStorage.ScheduleRule(ctx, ruleID, now)
	assert.NoError(t, err)

	scheduleKey := "rules:schedule"
	members, _ := mini.ZMembers(scheduleKey)
	assert.Contains(t, members, ruleID)
}

func TestRedisRuleStorage_GetScheduledRules(t *testing.T) {
	mini, redisClient := setupTestRedis(t)
	defer mini.Close()

	ruleStorage := NewRedisRuleStorage(redisClient)
	ctx := context.Background()

	now := time.Now().UTC()
	future := now.Add(1 * time.Hour)

	rule1 := "test-rule-1"
	rule2 := "test-rule-2"
	rule3 := "test-rule-3"

	err := ruleStorage.ScheduleRule(ctx, rule1, now.Add(-10*time.Minute))
	assert.NoError(t, err)
	err = ruleStorage.ScheduleRule(ctx, rule2, now)
	assert.NoError(t, err)
	err = ruleStorage.ScheduleRule(ctx, rule3, future)
	assert.NoError(t, err)

	rules, err := ruleStorage.GetScheduledRules(ctx, now)
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Contains(t, rules, rule1)
	assert.Contains(t, rules, rule2)
	assert.NotContains(t, rules, rule3)

	rules, err = ruleStorage.GetScheduledRules(ctx, future)
	assert.NoError(t, err)
	assert.Len(t, rules, 3)
}
