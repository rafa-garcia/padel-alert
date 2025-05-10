package testutil

import (
	"context"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

// MockMatchProcessor mocks the match processor interface
type MockMatchProcessor struct {
	mock.Mock
}

// Process mocks the match processing method
func (m *MockMatchProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	args := m.Called(ctx, rule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Activity), args.Error(1)
}
