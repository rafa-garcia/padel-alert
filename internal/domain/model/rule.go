package model

import (
	"time"
)

// Rule represents a notification rule for matches or classes
type Rule struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`   // Owner of the rule
	Type          string     `json:"rule_type"` // "match", "class", or "lesson"
	Name          string     `json:"name"`
	ClubIDs       []string   `json:"club_ids"`
	MinRanking    *float64   `json:"min_ranking,omitempty"`
	MaxRanking    *float64   `json:"max_ranking,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	TitleContains *string    `json:"title_contains,omitempty"`
	LastRun       *time.Time `json:"last_run,omitempty"` // Last execution time
	NextRun       *time.Time `json:"next_run,omitempty"` // Next scheduled run
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
