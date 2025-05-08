package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/padel-alert/internal/domain/model"
	"github.com/yourusername/padel-alert/internal/logger"
	"github.com/yourusername/padel-alert/internal/transformer"
	"github.com/yourusername/padel-alert/pkg/playtomic"
)

type ActivitySearchResponse struct {
	Count      int              `json:"count"`
	Activities []model.Activity `json:"activities"`
	Duration   string           `json:"duration"`
}

type SearchHandler struct {
	playtomicClient *playtomic.Client
}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		playtomicClient: playtomic.NewClient(),
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	query := r.URL.Query()
	clubIDs := parseClubIDs(query.Get("club_id"))

	dateParam := query.Get("date")
	var fromStartDate string
	if dateParam != "" {
		_, err := time.Parse("2006-01-02", dateParam)
		if err != nil {
			respondWithError(w, fmt.Sprintf("Invalid date format: %s", err.Error()), http.StatusBadRequest)
			return
		}
		fromStartDate = dateParam + "T00:00:00"
	} else {
		fromStartDate = time.Now().Format("2006-01-02") + "T00:00:00"
	}

	statusParam := query.Get("status")
	status := "PENDING,IN_PROGRESS"
	if statusParam != "" {
		validStatuses := map[string]bool{
			"PENDING":     true,
			"IN_PROGRESS": true,
			"PLAYED":      true,
			"FINISHED":    true,
		}

		statusValues := strings.Split(statusParam, ",")
		validStatusValues := []string{}

		for _, s := range statusValues {
			s = strings.TrimSpace(s)
			if validStatuses[s] {
				validStatusValues = append(validStatusValues, s)
			}
		}

		if len(validStatusValues) > 0 {
			status = strings.Join(validStatusValues, ",")
		}
	}

	playtomicTypeParam := query.Get("playtomic_type")
	classType := "COURSE,PUBLIC"
	if playtomicTypeParam != "" {
		validTypes := map[string]bool{
			"PRIVATE": true,
			"COURSE":  true,
			"PUBLIC":  true,
		}

		typeValues := strings.Split(playtomicTypeParam, ",")
		validTypeValues := []string{}

		for _, t := range typeValues {
			t = strings.TrimSpace(t)
			if validTypes[t] {
				validTypeValues = append(validTypeValues, t)
			}
		}

		if len(validTypeValues) > 0 {
			classType = strings.Join(validTypeValues, ",")
		}
	}

	pageSizeStr := query.Get("size")
	pageSize := 500
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			respondWithError(w, "Invalid size parameter", http.StatusBadRequest)
			return
		}
		if pageSize > 500 {
			pageSize = 500
		}
	}

	includeSummary := true
	if includeSummaryStr := query.Get("include_summary"); includeSummaryStr == "false" {
		includeSummary = false
	}

	includeUnavailable := false
	if includeUnavailableStr := query.Get("include_unavailable"); includeUnavailableStr == "true" {
		includeUnavailable = true
	}

	var minLevel float64
	minLevelStr := query.Get("min_level")
	filterByMinLevel := false
	if minLevelStr != "" {
		filterByMinLevel = true
		var err error
		minLevel, err = strconv.ParseFloat(minLevelStr, 64)
		if err != nil {
			respondWithError(w, "Invalid min_level parameter", http.StatusBadRequest)
			return
		}
	}

	var maxLevel float64
	maxLevelStr := query.Get("max_level")
	filterByMaxLevel := false
	if maxLevelStr != "" {
		filterByMaxLevel = true
		var err error
		maxLevel, err = strconv.ParseFloat(maxLevelStr, 64)
		if err != nil {
			respondWithError(w, "Invalid max_level parameter", http.StatusBadRequest)
			return
		}
	}

	searchQuery := query.Get("q")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	params := &playtomic.SearchClassesParams{
		Sort:             "start_date,created_at,ASC",
		Status:           status,
		Type:             classType,
		TenantIDs:        clubIDs,
		IncludeSummary:   includeSummary,
		Size:             pageSize,
		Page:             0,
		CourseVisibility: "PUBLIC",
		FromStartDate:    fromStartDate,
	}

	classes, err := h.playtomicClient.GetClasses(ctx, params)
	if err != nil {
		logger.Error("Error fetching classes", err)
		respondWithError(w, fmt.Sprintf("Error fetching data: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	activities, err := transformer.ClassesToActivities(classes)
	if err != nil {
		logger.Error("Error transforming classes", err)
		respondWithError(w, fmt.Sprintf("Error transforming data: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	filteredActivities := make([]model.Activity, 0)

	for _, activity := range activities {
		if !includeUnavailable && activity.AvailablePlaces <= 0 {
			continue
		}

		if filterByMinLevel {
			if activity.MinLevel < minLevel {
				continue
			}
		}

		if filterByMaxLevel {
			if activity.MaxLevel > maxLevel {
				continue
			}
		}

		filteredActivities = append(filteredActivities, activity)
	}

	if searchQuery != "" {
		searchQuery = strings.ToLower(searchQuery)
		searchResults := make([]model.Activity, 0)

		for _, activity := range filteredActivities {
			if strings.Contains(strings.ToLower(activity.Name), searchQuery) ||
				strings.Contains(strings.ToLower(activity.Club.Name), searchQuery) ||
				strings.Contains(strings.ToLower(activity.Club.Address.City), searchQuery) {
				searchResults = append(searchResults, activity)
			}
		}

		filteredActivities = searchResults
	}

	activityResp := ActivitySearchResponse{
		Count:      len(filteredActivities),
		Activities: filteredActivities,
		Duration:   time.Since(start).String(),
	}

	resp := Response{
		Data:   activityResp,
		Error:  nil,
		Status: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func parseClubIDs(clubIDsParam string) []string {
	if clubIDsParam == "" {
		return []string{}
	}
	return strings.Split(clubIDsParam, ",")
}
