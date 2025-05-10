package model

import (
	"time"
)

type Rule struct {
	ID         string    `json:"id"`
	Type       string    `json:"rule_type"`
	Name       string    `json:"name"`
	ClubIDs    []string  `json:"club_ids"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	Email      string    `json:"email"`
	TelegramID string    `json:"telegram_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	MinRanking *float64   `json:"min_ranking,omitempty"`
	MaxRanking *float64   `json:"max_ranking,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	TimeOfDay  []string   `json:"time_of_day,omitempty"`
	DaysOfWeek []string   `json:"days_of_week,omitempty"`

	TitleContains *string  `json:"title_contains,omitempty"`
	ClassTypes    []string `json:"class_types,omitempty"`

	LastChecked      time.Time `json:"last_checked,omitempty"`
	LastNotification time.Time `json:"last_notification,omitempty"`
	Active           bool      `json:"active"`
}

func NewRule(ruleType string, name string, clubIDs []string, userID string, userName string, email string) *Rule {
	now := time.Now()
	return &Rule{
		Type:      ruleType,
		Name:      name,
		ClubIDs:   clubIDs,
		UserID:    userID,
		UserName:  userName,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
		Active:    true,
	}
}

func NewMatchRule(name string, clubIDs []string, userID string, userName string, email string, minRanking, maxRanking float64, startDate, endDate time.Time) *Rule {
	rule := NewRule("match", name, clubIDs, userID, userName, email)
	rule.MinRanking = &minRanking
	rule.MaxRanking = &maxRanking

	startDateCopy := startDate
	endDateCopy := endDate
	rule.StartDate = &startDateCopy
	rule.EndDate = &endDateCopy
	return rule
}

func NewClassRule(name string, clubIDs []string, userID string, userName string, email string, titleContains string) *Rule {
	rule := NewRule("class", name, clubIDs, userID, userName, email)
	titleCopy := titleContains
	rule.TitleContains = &titleCopy
	return rule
}

func (r *Rule) IsMatch() bool {
	return r.Type == "match"
}

func (r *Rule) IsClass() bool {
	return r.Type == "class"
}
