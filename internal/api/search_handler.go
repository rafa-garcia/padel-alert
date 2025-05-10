package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/client"
	playtomicmodels "github.com/rafa-garcia/go-playtomic-api/models"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/transformer"
)

type ActivitySearchResponse struct {
	Count      int              `json:"count"`
	Activities []model.Activity `json:"activities"`
}

type SearchHandler struct {
	playtomicClient *client.Client
}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		playtomicClient: client.NewClient(
			client.WithTimeout(30*time.Second),
			client.WithRetries(3),
		),
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
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

	// Determine what types of activities to include based on type parameter
	activityTypeParam := query.Get("type")
	includeClasses := true
	includeTournaments := true
	includeCompetitiveMatches := true
	includeFriendlyMatches := true

	if activityTypeParam != "" {
		switch strings.ToUpper(activityTypeParam) {
		case "MATCH", "MATCH_COMPETITIVE", "MATCH_FRIENDLY":
			// For general match types
			if activityTypeParam == "MATCH" || activityTypeParam == "match" {
				includeClasses = false
				includeTournaments = false
			} else if strings.ToUpper(activityTypeParam) == "MATCH_COMPETITIVE" {
				includeClasses = false
				includeTournaments = false
				includeFriendlyMatches = false
			} else if strings.ToUpper(activityTypeParam) == "MATCH_FRIENDLY" {
				includeClasses = false
				includeTournaments = false
				includeCompetitiveMatches = false
			}
		case "CLASS":
			includeCompetitiveMatches = false
			includeFriendlyMatches = false
			includeTournaments = false
		case "TOURNAMENT":
			includeClasses = false
			includeCompetitiveMatches = false
			includeFriendlyMatches = false
		default:
			// If unrecognized type, include everything
			includeClasses = true
			includeTournaments = true
			includeCompetitiveMatches = true
			includeFriendlyMatches = true
		}
	}

	// For backward compatibility
	includeMatches := includeCompetitiveMatches || includeFriendlyMatches

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

	var allActivities []model.Activity

	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, 2)

	if includeClasses {
		wg.Add(1)
		go func() {
			defer wg.Done()

			classParams := &playtomicmodels.SearchClassesParams{
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

			classes, err := h.playtomicClient.GetClasses(ctx, classParams)
			if err != nil {
				logger.Error("Error fetching classes", err)
				errCh <- fmt.Errorf("error fetching class data: %w", err)
				return
			}

			classActivities, err := transformer.ExternalClassesToActivities(classes)
			if err != nil {
				logger.Error("Error transforming classes", err)
				errCh <- fmt.Errorf("error transforming class data: %w", err)
				return
			}

			mu.Lock()
			allActivities = append(allActivities, classActivities...)
			mu.Unlock()
		}()
	}

	if includeMatches {
		wg.Add(1)
		go func() {
			defer wg.Done()

			matchParams := &playtomicmodels.SearchMatchesParams{
				Sort:          "start_date,created_at,DESC",
				HasPlayers:    true,
				SportID:       "PADEL",
				TenantIDs:     clubIDs,
				Visibility:    "VISIBLE",
				FromStartDate: fromStartDate,
				Size:          pageSize,
				Page:          0,
			}

			matches, err := h.playtomicClient.GetMatches(ctx, matchParams)
			if err != nil {
				logger.Error("Error fetching matches", err)
				errCh <- fmt.Errorf("error fetching match data: %w", err)
				return
			}

			matchActivities, err := transformer.ExternalMatchesToActivities(matches)
			if err != nil {
				logger.Error("Error transforming matches", err)
				errCh <- fmt.Errorf("error transforming match data: %w", err)
				return
			}

			mu.Lock()
			allActivities = append(allActivities, matchActivities...)
			mu.Unlock()
		}()
	}

	if includeTournaments && len(clubIDs) > 0 {
		// For lessons, we need to make a separate request for each club ID since
		// the endpoint only accepts a single tenant_id
		for _, clubID := range clubIDs {
			wg.Add(1)
			go func(tenantID string) {
				defer wg.Done()

				lessonParams := &playtomicmodels.SearchLessonsParams{
					Sort:                 "start_date,created_at,ASC",
					TenantID:             tenantID,
					TournamentVisibility: "PUBLIC",
					Status:               "REGISTRATION_OPEN,REGISTRATION_CLOSED,IN_PROGRESS",
					Size:                 pageSize,
					Page:                 0,
					FromStartDate:        fromStartDate,
				}

				lessons, err := h.playtomicClient.GetLessons(ctx, lessonParams)
				if err != nil {
					logger.Error("Error fetching lessons", err, "tenantID", tenantID)
					errCh <- fmt.Errorf("error fetching lesson data for tenant %s: %w", tenantID, err)
					return
				}

				lessonActivities, err := transformer.ExternalLessonsToActivities(lessons)
				if err != nil {
					logger.Error("Error transforming lessons", err, "tenantID", tenantID)
					errCh <- fmt.Errorf("error transforming lesson data for tenant %s: %w", tenantID, err)
					return
				}

				mu.Lock()
				allActivities = append(allActivities, lessonActivities...)
				mu.Unlock()
			}(clubID)
		}
	}

	wg.Wait()

	select {
	case err := <-errCh:
		respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	default:
	}

	sortActivitiesByDate(allActivities)

	filteredActivities := make([]model.Activity, 0)

	for _, activity := range allActivities {
		// Filter based on activity type
		if strings.HasPrefix(activity.Type, "MATCH_") {
			if activity.Type == "MATCH_COMPETITIVE" && !includeCompetitiveMatches {
				continue
			}
			if activity.Type == "MATCH_FRIENDLY" && !includeFriendlyMatches {
				continue
			}
		} else if activity.Type == "TOURNAMENT" && !includeTournaments {
			continue
		} else if activity.Type == "ACADEMY_CLASS" && !includeClasses {
			continue
		}

		// Filter based on availability
		if !includeUnavailable && activity.AvailablePlaces <= 0 {
			continue
		}

		// Filter based on level
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

func sortActivitiesByDate(activities []model.Activity) {
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].StartDate.Before(activities[j].StartDate)
	})
}
