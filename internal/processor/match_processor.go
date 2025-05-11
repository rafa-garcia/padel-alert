package processor

import (
	"context"
	"fmt"
	"time"

	playtomic "github.com/rafa-garcia/go-playtomic-api/client"
	"github.com/rafa-garcia/go-playtomic-api/models"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/storage"
)

// MatchProcessor processes match rules
type MatchProcessor struct {
	client    *playtomic.Client
	ruleStore storage.RuleStorage
	redis     *storage.RedisClient
}

// NewMatchProcessor creates a new match processor
func NewMatchProcessor(client *playtomic.Client, ruleStore storage.RuleStorage, redis *storage.RedisClient) *MatchProcessor {
	return &MatchProcessor{
		client:    client,
		ruleStore: ruleStore,
		redis:     redis,
	}
}

// Process processes a match rule
func (p *MatchProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	if rule.Type != "match" {
		return nil, fmt.Errorf("not a match rule")
	}

	// Create search parameters
	params := &models.SearchMatchesParams{
		Sort:          "start_date,ASC",
		HasPlayers:    true,
		SportID:       "PADEL",
		TenantIDs:     rule.ClubIDs,
		Visibility:    "VISIBLE",
		FromStartDate: time.Now().Format("2006-01-02") + "T00:00:00",
		Size:          100,
		Page:          0,
	}

	// Fetch matches
	matches, err := p.client.GetMatches(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch matches: %w", err)
	}

	var activities []model.Activity

	// Process each match
	for _, m := range matches {
		// Skip matches without available spots
		if !hasAvailableSpots(m) {
			continue
		}

		// Apply ranking filter
		if !matchesRankingFilter(m, rule) {
			continue
		}

		// Apply date filter
		if !matchesDateFilter(m, rule) {
			continue
		}

		// Check if this match has been seen before
		if seen, _ := p.checkSeen(ctx, rule.ID, m.MatchID); seen {
			continue
		}

		// Add to activities
		activity := convertMatchToActivity(m)
		activities = append(activities, activity)

		// Mark as seen
		if err := p.markSeen(ctx, rule.ID, m.MatchID); err != nil {
			logger.Error("Failed to mark match as seen", err, "rule_id", rule.ID, "match_id", m.MatchID)
		}
	}

	return activities, nil
}

// hasAvailableSpots checks if the match has available spots
func hasAvailableSpots(m models.Match) bool {
	currentPlayers := 0
	maxPlayers := 0

	for _, team := range m.Teams {
		currentPlayers += len(team.Players)
		maxPlayers += m.MaxPlayersPerTeam
	}

	return currentPlayers < maxPlayers
}

// matchesRankingFilter checks if the match matches the rule's ranking filter
func matchesRankingFilter(m models.Match, rule *model.Rule) bool {
	if rule.MinRanking != nil && m.MinLevel < *rule.MinRanking {
		return false
	}

	if rule.MaxRanking != nil && m.MaxLevel > *rule.MaxRanking {
		return false
	}

	return true
}

// matchesDateFilter checks if the match matches the rule's date filter
func matchesDateFilter(m models.Match, rule *model.Rule) bool {
	matchTime, err := time.Parse("2006-01-02T15:04:05", m.StartDate)
	if err != nil {
		logger.Error("Failed to parse match date", err, "date", m.StartDate)
		return false
	}

	if rule.StartDate != nil && matchTime.Before(*rule.StartDate) {
		return false
	}

	if rule.EndDate != nil && matchTime.After(*rule.EndDate) {
		return false
	}

	return true
}

// checkSeen checks if a match has been seen before for a rule
func (p *MatchProcessor) checkSeen(ctx context.Context, ruleID string, matchID string) (bool, error) {
	key := fmt.Sprintf("seen:%s", ruleID)
	return p.redis.Client.SIsMember(ctx, key, matchID).Result()
}

// markSeen marks a match as seen for a rule
func (p *MatchProcessor) markSeen(ctx context.Context, ruleID string, matchID string) error {
	key := fmt.Sprintf("seen:%s", ruleID)
	return p.redis.Client.SAdd(ctx, key, matchID).Err()
}

// convertMatchToActivity converts a Playtomic match to an Activity
func convertMatchToActivity(m models.Match) model.Activity {
	// Calculate available places
	currentPlayers := 0
	maxPlayers := 0

	for _, team := range m.Teams {
		currentPlayers += len(team.Players)
		maxPlayers += m.MaxPlayersPerTeam
	}

	matchType := "MATCH_COMPETITIVE"
	if m.MatchType == "FRIENDLY" {
		matchType = "MATCH_FRIENDLY"
	}

	startTime, _ := time.Parse("2006-01-02T15:04:05", m.StartDate)
	endTime, _ := time.Parse("2006-01-02T15:04:05", m.EndDate)

	// Calculate duration in minutes
	duration := int(endTime.Sub(startTime).Minutes())

	// Convert players
	players := make([]model.Player, 0)
	for _, team := range m.Teams {
		teamStr := team.TeamID
		for _, p := range team.Players {
			players = append(players, model.Player{
				ID:    p.UserID,
				Name:  p.Name,
				Level: p.LevelValue,
				Team:  &teamStr,
				Link:  fmt.Sprintf("https://playtomic.io/user/%s", p.UserID),
			})
		}
	}

	return model.Activity{
		ID:                m.MatchID,
		Type:              matchType,
		PlaytomicType:     m.MatchType,
		Name:              fmt.Sprintf("%s Match", m.MatchType),
		StartDate:         startTime,
		EndDate:           endTime,
		Duration:          duration,
		MinPlayers:        m.MinPlayersPerTeam * len(m.Teams),
		MaxPlayers:        maxPlayers,
		MinLevel:          m.MinLevel,
		MaxLevel:          m.MaxLevel,
		Price:             m.Price,
		Gender:            m.Gender,
		AvailablePlaces:   maxPlayers - currentPlayers,
		RegisteredPlayers: players,
		Link:              fmt.Sprintf("https://playtomic.io/match/%s", m.MatchID),
		Club: model.Club{
			ID:   m.Tenant.TenantID,
			Name: m.Tenant.TenantName,
			Address: model.Address{
				Street:     m.Tenant.Address.Street,
				PostalCode: m.Tenant.Address.PostalCode,
				City:       m.Tenant.Address.City,
				Country:    m.Tenant.Address.Country,
			},
			Link: fmt.Sprintf("https://playtomic.io/club/%s", m.Tenant.TenantID),
		},
	}
}
