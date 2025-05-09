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

// CreateRuleRequest represents a request to create a new rule
type CreateRuleRequest struct {
	Type          string     `json:"rule_type"` // "match", "class", or "lesson"
	Name          string     `json:"name"`
	ClubIDs       []string   `json:"club_ids"`
	MinRanking    *float64   `json:"min_ranking,omitempty"`
	MaxRanking    *float64   `json:"max_ranking,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	TitleContains *string    `json:"title_contains,omitempty"`
}

// UpdateRuleRequest represents a request to update an existing rule
type UpdateRuleRequest struct {
	Name          string     `json:"name"`
	ClubIDs       []string   `json:"club_ids"`
	MinRanking    *float64   `json:"min_ranking,omitempty"`
	MaxRanking    *float64   `json:"max_ranking,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	TitleContains *string    `json:"title_contains,omitempty"`
}

// ListRules lists all rules for a user
func (h *RuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user ID
	userID := r.Context().Value(contextKeyUserID).(string)

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
	// Get the authenticated user ID
	userID := r.Context().Value(contextKeyUserID).(string)

	// Get the rule ID from the path
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

	// Check ownership
	if rule.UserID != userID {
		respondWithError(w, "Not authorized to access this rule", http.StatusForbidden)
		return
	}

	respondWithJSON(w, rule)
}

// CreateRule creates a new rule
func (h *RuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user ID
	userID := r.Context().Value(contextKeyUserID).(string)

	// Decode request
	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" || len(req.ClubIDs) == 0 {
		respondWithError(w, "Name and club_ids are required", http.StatusBadRequest)
		return
	}

	// Validate rule type
	if req.Type != "match" && req.Type != "class" && req.Type != "lesson" {
		respondWithError(w, "Invalid rule_type: must be 'match', 'class', or 'lesson'", http.StatusBadRequest)
		return
	}

	// Create rule
	rule := &model.Rule{
		ID:            util.GenerateID(),
		UserID:        userID,
		Type:          req.Type,
		Name:          req.Name,
		ClubIDs:       req.ClubIDs,
		MinRanking:    req.MinRanking,
		MaxRanking:    req.MaxRanking,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		TitleContains: req.TitleContains,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save rule
	if err := h.ruleStorage.CreateRule(r.Context(), rule); err != nil {
		logger.Error("Failed to create rule", err)
		respondWithError(w, "Failed to create rule", http.StatusInternalServerError)
		return
	}

	// Schedule rule for immediate processing
	if err := h.ruleStorage.ScheduleRule(r.Context(), rule.ID, time.Now()); err != nil {
		logger.Error("Failed to schedule rule", err, "rule_id", rule.ID)
		// Don't return error to client, just log it
	}

	respondWithJSON(w, rule)
}

// UpdateRule updates an existing rule
func (h *RuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user ID
	userID := r.Context().Value(contextKeyUserID).(string)

	// Get the rule ID from the path
	ruleID := chi.URLParam(r, "id")
	if ruleID == "" {
		respondWithError(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	// Decode request
	var req UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get existing rule
	rule, err := h.ruleStorage.GetRule(r.Context(), ruleID)
	if err != nil {
		logger.Error("Failed to get rule", err, "rule_id", ruleID)
		respondWithError(w, "Rule not found", http.StatusNotFound)
		return
	}

	// Check ownership
	if rule.UserID != userID {
		respondWithError(w, "Not authorized to update this rule", http.StatusForbidden)
		return
	}

	// Update rule
	rule.Name = req.Name
	rule.ClubIDs = req.ClubIDs
	rule.MinRanking = req.MinRanking
	rule.MaxRanking = req.MaxRanking
	rule.StartDate = req.StartDate
	rule.EndDate = req.EndDate
	rule.TitleContains = req.TitleContains
	rule.UpdatedAt = time.Now()

	// Save updated rule
	if err := h.ruleStorage.UpdateRule(r.Context(), rule); err != nil {
		logger.Error("Failed to update rule", err, "rule_id", ruleID)
		respondWithError(w, "Failed to update rule", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, rule)
}

// DeleteRule deletes a rule
func (h *RuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user ID
	userID := r.Context().Value(contextKeyUserID).(string)

	// Get the rule ID from the path
	ruleID := chi.URLParam(r, "id")
	if ruleID == "" {
		respondWithError(w, "Rule ID is required", http.StatusBadRequest)
		return
	}

	// Get existing rule
	rule, err := h.ruleStorage.GetRule(r.Context(), ruleID)
	if err != nil {
		logger.Error("Failed to get rule", err, "rule_id", ruleID)
		respondWithError(w, "Rule not found", http.StatusNotFound)
		return
	}

	// Check ownership
	if rule.UserID != userID {
		respondWithError(w, "Not authorized to delete this rule", http.StatusForbidden)
		return
	}

	// Delete rule
	if err := h.ruleStorage.DeleteRule(r.Context(), ruleID); err != nil {
		logger.Error("Failed to delete rule", err, "rule_id", ruleID)
		respondWithError(w, "Failed to delete rule", http.StatusInternalServerError)
		return
	}

	respondWithSuccess(w, "Rule deleted successfully")
}
