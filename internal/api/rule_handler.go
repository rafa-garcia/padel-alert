package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/storage"
	"github.com/rafa-garcia/padel-alert/internal/util"
)

// RuleHandler handles API requests for rules
type RuleHandler struct {
	ruleStorage storage.RuleStorage
	userStorage storage.UserStorage
}

// NewRuleHandler creates a new rule handler
func NewRuleHandler(ruleStorage storage.RuleStorage, userStorage storage.UserStorage) *RuleHandler {
	return &RuleHandler{
		ruleStorage: ruleStorage,
		userStorage: userStorage,
	}
}

type CreateRuleRequest struct {
	Type          string   `json:"rule_type"`
	Name          string   `json:"name"`
	ClubIDs       []string `json:"club_ids"`
	UserID        string   `json:"user_id"`
	UserName      string   `json:"user_name"`
	Email         string   `json:"email"`
	MinRanking    *float64 `json:"min_ranking,omitempty"`
	MaxRanking    *float64 `json:"max_ranking,omitempty"`
	StartDate     string   `json:"start_date,omitempty"`
	EndDate       string   `json:"end_date,omitempty"`
	TitleContains *string  `json:"title_contains,omitempty"`
}

// UpdateRuleRequest represents a request to update an existing rule
type UpdateRuleRequest struct {
	Name          string     `json:"name"`
	UserName      string     `json:"user_name"`
	ClubIDs       []string   `json:"club_ids"`
	MinRanking    *float64   `json:"min_ranking,omitempty"`
	MaxRanking    *float64   `json:"max_ranking,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	TitleContains *string    `json:"title_contains,omitempty"`
}

// ListRules lists all rules for a user
func (h *RuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	// Try to get the authenticated user ID from context, but don't fail if it's missing
	contextUserID, ok := GetUserID(r.Context())
	if !ok {
		respondWithError(w, "User ID not found in request context", http.StatusUnauthorized)
		return
	}

	userID := contextUserID

	rules, err := h.ruleStorage.ListRules(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to list rules", err, "user_id", userID)
		respondWithError(w, "Failed to list rules", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, rules)
}

// GetRule gets a specific rule
func (h *RuleHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	// Try to get the authenticated user ID from context
	contextUserID, ok := GetUserID(r.Context())
	if !ok {
		respondWithError(w, "User ID not found in request context", http.StatusUnauthorized)
		return
	}

	userID := contextUserID

	ruleID := chi.URLParam(r, "id")
	if ruleID == "" {
		respondWithError(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	rule, err := h.ruleStorage.GetRule(r.Context(), ruleID)
	if err != nil {
		logger.Error("Failed to get rule", err, "rule_id", ruleID)
		respondWithError(w, "Rule not found", http.StatusNotFound)
		return
	}

	if rule.UserID != userID {
		respondWithError(w, "Not authorized to access this rule", http.StatusForbidden)
		return
	}

	respondWithJSON(w, rule)
}

// CreateRule creates a new rule
func (h *RuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var requestUserID string

	contextUserID, ok := GetUserID(r.Context())
	if ok {
		requestUserID = contextUserID
	}

	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" || len(req.ClubIDs) == 0 {
		respondWithError(w, "Name and club_ids are required", http.StatusBadRequest)
		return
	}

	if req.Type != "match" && req.Type != "class" && req.Type != "lesson" {
		respondWithError(w, "Invalid rule_type: must be 'match', 'class', or 'lesson'", http.StatusBadRequest)
		return
	}

	if req.UserID == "" && requestUserID == "" {
		respondWithError(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		respondWithError(w, "email is required", http.StatusBadRequest)
		return
	}

	effectiveUserID := requestUserID
	if effectiveUserID == "" {
		effectiveUserID = req.UserID
	}

	var startDate, endDate *time.Time

	if req.StartDate != "" {
		parsedTime, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			respondWithError(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startDate = &parsedTime
	}

	if req.EndDate != "" {
		parsedTime, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			respondWithError(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		endDate = &parsedTime
	}

	// Check for username
	if req.UserName == "" {
		req.UserName = effectiveUserID // Use user ID as fallback
	}

	rule := &model.Rule{
		ID:            util.GenerateID(),
		UserID:        effectiveUserID,
		UserName:      req.UserName,
		Email:         req.Email,
		Type:          req.Type,
		Name:          req.Name,
		ClubIDs:       req.ClubIDs,
		MinRanking:    req.MinRanking,
		MaxRanking:    req.MaxRanking,
		StartDate:     startDate,
		EndDate:       endDate,
		TitleContains: req.TitleContains,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Active:        true, // Set rules to active by default
	}

	if err := h.ruleStorage.CreateRule(r.Context(), rule); err != nil {
		logger.Error("Failed to create rule", err)
		respondWithError(w, "Failed to create rule", http.StatusInternalServerError)
		return
	}

	if err := h.ruleStorage.ScheduleRule(r.Context(), rule.ID, time.Now()); err != nil {
		logger.Error("Failed to schedule rule", err, "rule_id", rule.ID)
		// Don't return error to client, just log it
	}

	w.WriteHeader(http.StatusCreated)
	respondWithJSON(w, rule)
}

// UpdateRule updates an existing rule
func (h *RuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	contextUserID, ok := GetUserID(r.Context())
	if !ok {
		respondWithError(w, "User ID not found in request context", http.StatusUnauthorized)
		return
	}

	userID := contextUserID

	ruleID := chi.URLParam(r, "id")
	if ruleID == "" {
		respondWithError(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	var req UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	rule, err := h.ruleStorage.GetRule(r.Context(), ruleID)
	if err != nil {
		logger.Error("Failed to get rule", err, "rule_id", ruleID)
		respondWithError(w, "Rule not found", http.StatusNotFound)
		return
	}

	if rule.UserID != userID {
		respondWithError(w, "Not authorized to update this rule", http.StatusForbidden)
		return
	}

	rule.Name = req.Name
	if req.UserName != "" {
		rule.UserName = req.UserName
	}
	rule.ClubIDs = req.ClubIDs
	rule.MinRanking = req.MinRanking
	rule.MaxRanking = req.MaxRanking
	rule.StartDate = req.StartDate
	rule.EndDate = req.EndDate
	rule.TitleContains = req.TitleContains
	rule.UpdatedAt = time.Now()

	if err := h.ruleStorage.UpdateRule(r.Context(), rule); err != nil {
		logger.Error("Failed to update rule", err, "rule_id", ruleID)
		respondWithError(w, "Failed to update rule", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, rule)
}

// DeleteRule deletes a rule
func (h *RuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	contextUserID, ok := GetUserID(r.Context())
	if !ok {
		respondWithError(w, "User ID not found in request context", http.StatusUnauthorized)
		return
	}

	userID := contextUserID

	ruleID := chi.URLParam(r, "id")
	if ruleID == "" {
		respondWithError(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	rule, err := h.ruleStorage.GetRule(r.Context(), ruleID)
	if err != nil {
		logger.Error("Failed to get rule", err, "rule_id", ruleID)
		respondWithError(w, "Rule not found", http.StatusNotFound)
		return
	}

	if rule.UserID != userID {
		respondWithError(w, "Not authorized to delete this rule", http.StatusForbidden)
		return
	}

	if err := h.ruleStorage.DeleteRule(r.Context(), ruleID); err != nil {
		logger.Error("Failed to delete rule", err, "rule_id", ruleID)
		respondWithError(w, "Failed to delete rule", http.StatusInternalServerError)
		return
	}

	respondWithSuccess(w, "Rule deleted successfully")
}
