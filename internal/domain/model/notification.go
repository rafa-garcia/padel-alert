package model

import (
	"time"
)

// Notification represents a sent or pending notification
type Notification struct {
	ID        string     `json:"id"`
	RuleID    string     `json:"rule_id"`
	Message   string     `json:"message"`
	Sent      bool       `json:"sent"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
