package testutil

import (
	"context"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

// MockProcessor mocks the generic processor interface
type MockProcessor struct {
	mock.Mock
}

// Process mocks the processing method
func (m *MockProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	args := m.Called(ctx, rule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Activity), args.Error(1)
}

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

// MockClassProcessor mocks the class processor interface
type MockClassProcessor struct {
	mock.Mock
}

// Process mocks the class processing method
func (m *MockClassProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	args := m.Called(ctx, rule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Activity), args.Error(1)
}

// MockLessonProcessor mocks the lesson processor interface
type MockLessonProcessor struct {
	mock.Mock
}

// Process mocks the lesson processing method
func (m *MockLessonProcessor) Process(ctx context.Context, rule *model.Rule) ([]model.Activity, error) {
	args := m.Called(ctx, rule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Activity), args.Error(1)
}
