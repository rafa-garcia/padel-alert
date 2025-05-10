package notification

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

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

	if n.config.SMTPServer == "" || n.config.SMTPUsername == "" || n.config.SMTPPassword == "" {
		logger.Warn("SMTP not configured, skipping email notification", "user_id", user.ID)
		return nil
	}

	subject := fmt.Sprintf("PadelAlert: %d new activities available", len(activities))
	htmlBody, err := n.formatEmailHTML(rule, activities)
	if err != nil {
		return fmt.Errorf("format email: %w", err)
	}

	err = n.sendEmail(user.Email, subject, htmlBody)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	logger.Info("Email notification sent", "user_id", user.ID, "email", user.Email, "activities", len(activities))
	return nil
}

// sendEmail sends an email via SMTP
func (n *EmailNotifier) sendEmail(to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", n.config.SMTPUsername, n.config.SMTPPassword, n.config.SMTPServer)

	headers := map[string]string{
		"From":         n.config.SMTPSender,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=utf-8",
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	addr := fmt.Sprintf("%s:%d", n.config.SMTPServer, n.config.SMTPPort)
	return smtp.SendMail(addr, auth, n.config.SMTPSender, []string{to}, []byte(message))
}

// formatEmailHTML formats the email body as HTML using a template file
func (n *EmailNotifier) formatEmailHTML(rule *model.Rule, activities []model.Activity) (string, error) {
	funcMap := template.FuncMap{
		"len": func(items []model.Activity) int {
			return len(items)
		},
		"formatDate": func(t interface{}) string {
			if timeVal, ok := t.(time.Time); ok {
				return timeVal.Format("2006-01-02")
			}
			return fmt.Sprintf("%v", t)
		},
		"formatTime": func(t interface{}) string {
			if timeVal, ok := t.(time.Time); ok {
				return timeVal.Format("15:04")
			}
			return fmt.Sprintf("%v", t)
		},
		"formatDuration": func(start, end interface{}) string {
			startTime, startOk := start.(time.Time)
			endTime, endOk := end.(time.Time)

			if !startOk || !endOk {
				return ""
			}

			duration := endTime.Sub(startTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60

			if minutes == 0 {
				return fmt.Sprintf("%dh", hours)
			}
			return fmt.Sprintf("%dh %dm", hours, minutes)
		},
		"formatLevel": func(level interface{}) string {
			if floatVal, ok := level.(float64); ok {
				if floatVal == 0 {
					return "Any"
				}
				return fmt.Sprintf("%.1f", floatVal)
			}
			return fmt.Sprintf("%v", level)
		},
	}

	cwd, _ := os.Getwd()
	templateFile := "activity_notification.html"
	templatePath := filepath.Join(cwd, "assets", "templates", templateFile)

	tmpl, err := template.New(templateFile).Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("parse template file: %w", err)
	}

	var buf bytes.Buffer
	userData := &model.User{
		ID:    rule.UserID,
		Email: rule.Email,
	}

	if rule.UserName != "" {
		userData.Name = rule.UserName
	}

	data := map[string]interface{}{
		"Rule":       rule,
		"Activities": activities,
		"User":       userData,
	}

	err = tmpl.ExecuteTemplate(&buf, templateFile, data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
