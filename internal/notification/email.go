package notification

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
)

// EmailNotifier handles email notifications
type EmailNotifier struct {
	config *config.Config
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(cfg *config.Config) *EmailNotifier {
	return &EmailNotifier{
		config: cfg,
	}
}

// NotifyNewActivities sends notifications about new activities
func (n *EmailNotifier) NotifyNewActivities(ctx context.Context, user *model.User, rule *model.Rule, activities []model.Activity) error {
	if len(activities) == 0 {
		return nil
	}

	// If SMTP settings are not configured, log but don't return error
	if n.config.SMTPServer == "" || n.config.SMTPUsername == "" || n.config.SMTPPassword == "" {
		logger.Warn("SMTP not configured, skipping email notification", "user_id", user.ID)
		return nil
	}

	// Format email content
	subject := fmt.Sprintf("PadelAlert: %d new activities available", len(activities))
	htmlBody, err := n.formatEmailHTML(rule, activities)
	if err != nil {
		return fmt.Errorf("format email: %w", err)
	}

	// Send email
	err = n.sendEmail(user.Email, subject, htmlBody)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	logger.Info("Email notification sent", "user_id", user.ID, "email", user.Email, "activities", len(activities))
	return nil
}

// sendEmail sends an email
func (n *EmailNotifier) sendEmail(to, subject, htmlBody string) error {
	// Set up authentication
	auth := smtp.PlainAuth("", n.config.SMTPUsername, n.config.SMTPPassword, n.config.SMTPServer)

	// Set up headers
	headers := map[string]string{
		"From":         n.config.SMTPSender,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=utf-8",
	}

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Send email
	addr := fmt.Sprintf("%s:%d", n.config.SMTPServer, n.config.SMTPPort)
	return smtp.SendMail(addr, auth, n.config.SMTPSender, []string{to}, []byte(message))
}

// formatEmailHTML formats the email body as HTML
func (n *EmailNotifier) formatEmailHTML(rule *model.Rule, activities []model.Activity) (string, error) {
	// Simple HTML template
	const emailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .activity { margin-bottom: 20px; border-bottom: 1px solid #eee; padding-bottom: 10px; }
        .header { font-weight: bold; }
        .details { margin-left: 10px; }
    </style>
</head>
<body>
    <h2>PadelAlert: New Padel Activities Available</h2>
    <p>Your rule "{{.Rule.Name}}" found {{len .Activities}} new activities:</p>
    
    {{range .Activities}}
    <div class="activity">
        <div class="header">{{.Name}} at {{.Club.Name}}</div>
        <div class="details">
            <p>📅 Date: {{formatTime .StartDate}} - {{formatTime .EndDate}}</p>
            <p>🌟 Level: {{.MinLevel}} - {{.MaxLevel}}</p>
            <p>👥 Available places: {{.AvailablePlaces}}</p>
            <p>💰 Price: {{.Price}}</p>
            <p><a href="{{.Link}}">View on Playtomic</a></p>
        </div>
    </div>
    {{end}}
    
    <p>Happy playing!</p>
</body>
</html>
`

	// Create template with functions
	tmpl, err := template.New("email").Funcs(template.FuncMap{
		"len": func(items []model.Activity) int {
			return len(items)
		},
		"formatTime": func(t interface{}) string {
			if t, ok := t.(fmt.Stringer); ok {
				return t.String()
			}
			return fmt.Sprintf("%v", t)
		},
	}).Parse(emailTemplate)

	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Rule":       rule,
		"Activities": activities,
	})

	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
