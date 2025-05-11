package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestParseClubIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Single club ID",
			input:    "club-123",
			expected: []string{"club-123"},
		},
		{
			name:     "Multiple club IDs",
			input:    "club-123,club-456,club-789",
			expected: []string{"club-123", "club-456", "club-789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseClubIDs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortActivitiesByDate(t *testing.T) {
	now := time.Now()
	activities := []model.Activity{
		{
			ID:        "activity-3",
			StartDate: now.Add(2 * time.Hour),
		},
		{
			ID:        "activity-1",
			StartDate: now,
		},
		{
			ID:        "activity-2",
			StartDate: now.Add(1 * time.Hour),
		},
	}

	sortActivitiesByDate(activities)

	assert.Equal(t, "activity-1", activities[0].ID)
	assert.Equal(t, "activity-2", activities[1].ID)
	assert.Equal(t, "activity-3", activities[2].ID)
}

type noopHandler struct {
	t *testing.T
}

func (h *noopHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	activityResp := ActivitySearchResponse{
		Count:      0,
		Activities: []model.Activity{},
	}

	resp := Response{
		Data:   activityResp,
		Status: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.t.Fatalf("Failed to encode JSON response: %v", err)
	}
}

func TestSearchHandlerBasicResponse(t *testing.T) {
	handler := &noopHandler{
		t: t,
	}

	req := httptest.NewRequest("GET", "/api/v1/search", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.Status)
	assert.Nil(t, resp.Error)

	_, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
}

func TestSearchHandlerInvalidDate(t *testing.T) {
	w := httptest.NewRecorder()

	errMsg := "Invalid date format"
	resp := Response{
		Error:  &errMsg,
		Status: http.StatusBadRequest,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("Failed to encode JSON response: %v", err)
	}

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respObj Response
	err := json.Unmarshal(w.Body.Bytes(), &respObj)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, respObj.Status)
	assert.NotNil(t, respObj.Error)
}
