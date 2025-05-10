package testutil

import (
	"context"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

// MockEmailNotifier mocks the email notifier interface
type MockEmailNotifier struct {
	mock.Mock
}

// NotifyNewActivities mocks the email notification method
func (m *MockEmailNotifier) NotifyNewActivities(ctx context.Context, user *model.User, rule *model.Rule, activities []model.Activity) error {
	if len(activities) > 0 && rule != nil {
		return m.Called(ctx, user, rule, activities).Error(0)
	}
	return m.Called(ctx, user, rule, activities).Error(0)
}
