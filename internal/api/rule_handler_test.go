package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRuleStorage struct {
	mock.Mock
}

func (m *MockRuleStorage) GetRule(ctx context.Context, ruleID string) (*model.Rule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rule), args.Error(1)
}

func (m *MockRuleStorage) ListRules(ctx context.Context, userID string) ([]*model.Rule, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Rule), args.Error(1)
}

func (m *MockRuleStorage) CreateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleStorage) UpdateRule(ctx context.Context, rule *model.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleStorage) DeleteRule(ctx context.Context, ruleID string) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockRuleStorage) ScheduleRule(ctx context.Context, ruleID string, nextRun time.Time) error {
	args := m.Called(ctx, ruleID, nextRun)
	return args.Error(0)
}

func (m *MockRuleStorage) GetScheduledRules(ctx context.Context, until time.Time) ([]string, error) {
	args := m.Called(ctx, until)
	return args.Get(0).([]string), args.Error(1)
}

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) GetUser(ctx context.Context, userID string) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStorage) CreateUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserStorage) UpdateUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserStorage) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestRuleHandler_ListRules(t *testing.T) {
	ruleStorage := new(MockRuleStorage)
	userStorage := new(MockUserStorage)
	handler := NewRuleHandler(ruleStorage, userStorage)

	userID := "test-user-123"
	rules := []*model.Rule{
		{
			ID:       "rule-1",
			UserID:   userID,
			UserName: "Test User",
			Name:     "Rule 1",
		},
		{
			ID:       "rule-2",
			UserID:   userID,
			UserName: "Test User",
			Name:     "Rule 2",
		},
	}

	ruleStorage.On("ListRules", mock.Anything, userID).Return(rules, nil)

	req := httptest.NewRequest("GET", "/api/v1/rules", nil)
	req = req.WithContext(WithUserID(req.Context(), userID))

	w := httptest.NewRecorder()
	handler.ListRules(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	ruleStorage.AssertExpectations(t)
}

func TestRuleHandler_GetRule(t *testing.T) {
	ruleStorage := new(MockRuleStorage)
	userStorage := new(MockUserStorage)
	handler := NewRuleHandler(ruleStorage, userStorage)

	userID := "test-user-123"
	ruleID := "rule-1"

	rule := &model.Rule{
		ID:       ruleID,
		UserID:   userID,
		UserName: "Test User",
		Name:     "Test Rule",
	}

	ruleStorage.On("GetRule", mock.Anything, ruleID).Return(rule, nil)

	r := chi.NewRouter()
	r.Get("/{id}", handler.GetRule)

	req := httptest.NewRequest("GET", "/"+ruleID, nil)
	req = req.WithContext(WithUserID(req.Context(), userID))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	ruleStorage.AssertExpectations(t)
}

func TestRuleHandler_CreateRule(t *testing.T) {
	ruleStorage := new(MockRuleStorage)
	userStorage := new(MockUserStorage)
	handler := NewRuleHandler(ruleStorage, userStorage)

	userID := "test-user-123"

	createReq := CreateRuleRequest{
		Type:     "match",
		Name:     "Test Rule",
		ClubIDs:  []string{"club-1"},
		UserName: "Test User",
		Email:    "test@example.com",
	}

	body, _ := json.Marshal(createReq)

	ruleStorage.On("CreateRule", mock.Anything, mock.MatchedBy(func(rule *model.Rule) bool {
		return rule.Name == createReq.Name &&
			rule.UserID == userID &&
			rule.UserName == createReq.UserName
	})).Return(nil)

	ruleStorage.On("ScheduleRule", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/rules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(WithUserID(req.Context(), userID))

	w := httptest.NewRecorder()
	handler.CreateRule(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	ruleStorage.AssertExpectations(t)
}

func TestRuleHandler_DeleteRule(t *testing.T) {
	ruleStorage := new(MockRuleStorage)
	userStorage := new(MockUserStorage)
	handler := NewRuleHandler(ruleStorage, userStorage)

	userID := "test-user-123"
	ruleID := "rule-1"

	rule := &model.Rule{
		ID:       ruleID,
		UserID:   userID,
		UserName: "Test User",
		Name:     "Test Rule",
	}

	ruleStorage.On("GetRule", mock.Anything, ruleID).Return(rule, nil)
	ruleStorage.On("DeleteRule", mock.Anything, ruleID).Return(nil)

	r := chi.NewRouter()
	r.Delete("/{id}", handler.DeleteRule)

	req := httptest.NewRequest("DELETE", "/"+ruleID, nil)
	req = req.WithContext(WithUserID(req.Context(), userID))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	ruleStorage.AssertExpectations(t)
}
