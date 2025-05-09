package notification

import (
	"context"
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestEmailNotifier_FormatEmailHTML(t *testing.T) {
	cfg := &config.Config{
		SMTPServer:   "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		SMTPSender:   "noreply@example.com",
	}

	notifier := NewEmailNotifier(cfg)

	rule := &model.Rule{
		ID:   "rule-1",
		Name: "Test Rule",
		Type: "match",
	}

	now := time.Now()
	activities := []model.Activity{
		{
			ID:              "activity-1",
			Type:            "MATCH_COMPETITIVE",
			Name:            "Competitive Match",
			StartDate:       now,
			EndDate:         now.Add(90 * time.Minute),
			MinLevel:        3.0,
			MaxLevel:        4.0,
			Price:           "€20",
			AvailablePlaces: 2,
			Club: model.Club{
				ID:   "club-1",
				Name: "Test Club",
				Address: model.Address{
					City:    "Madrid",
					Country: "Spain",
				},
			},
			Link: "https://example.com/match",
		},
	}

	html, err := notifier.formatEmailHTML(rule, activities)
	assert.NoError(t, err)
	assert.Contains(t, html, "PadelAlert: New Padel Activities Available")
	assert.Contains(t, html, "Your rule \"Test Rule\"")
	assert.Contains(t, html, "Competitive Match at Test Club")
	assert.Contains(t, html, "Level: 3 - 4")
	assert.Contains(t, html, "Available places: 2")
	assert.Contains(t, html, "€20")
}

func TestEmailNotifier_NotifyNewActivities_NoActivities(t *testing.T) {
	cfg := &config.Config{}
	notifier := NewEmailNotifier(cfg)

	user := &model.User{
		ID:    "user-1",
		Email: "user@example.com",
	}

	rule := &model.Rule{
		ID:   "rule-1",
		Name: "Test Rule",
	}

	err := notifier.NotifyNewActivities(context.Background(), user, rule, []model.Activity{})
	assert.NoError(t, err)
}

func TestEmailNotifier_NotifyNewActivities_NoSMTPConfig(t *testing.T) {
	cfg := &config.Config{} // No SMTP configuration
	notifier := NewEmailNotifier(cfg)

	user := &model.User{
		ID:    "user-1",
		Email: "user@example.com",
	}

	rule := &model.Rule{
		ID:   "rule-1",
		Name: "Test Rule",
	}

	activities := []model.Activity{
		{
			ID:   "activity-1",
			Name: "Test Activity",
		},
	}

	err := notifier.NotifyNewActivities(context.Background(), user, rule, activities)
	assert.NoError(t, err)
}
