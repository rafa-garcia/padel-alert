package transformer

import (
	"fmt"
	"math"

	"github.com/rafa-garcia/go-playtomic-api/models"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
)

// ExternalLessonToActivity transforms an external Playtomic lesson to our domain Activity model
func ExternalLessonToActivity(lesson models.Lesson) (model.Activity, error) {
	activity := model.Activity{
		ID:                lesson.TournamentID,
		Type:              "TOURNAMENT", // Identifying these as tournaments in our system
		PlaytomicType:     lesson.Type,
		RegisteredPlayers: []model.Player{},
		Name:              lesson.TournamentName,
		Gender:            lesson.Gender,
		MinPlayers:        lesson.MinPlayers,
		MaxPlayers:        lesson.MaxPlayers,
		AvailablePlaces:   lesson.AvailablePlaces,
		Price:             lesson.Price,
	}

	startDate, err := models.ParseTime(lesson.StartDate)
	if err != nil {
		return activity, fmt.Errorf("parsing start date: %w", err)
	}
	activity.StartDate = startDate

	endDate, err := models.ParseTime(lesson.EndDate)
	if err != nil {
		return activity, fmt.Errorf("parsing end date: %w", err)
	}
	activity.EndDate = endDate

	duration := endDate.Sub(startDate).Minutes()
	activity.Duration = int(duration)

	// Convert tenantAddress to our address model
	activity.Club = model.Club{
		ID:   lesson.Tenant.TenantID,
		Name: lesson.Tenant.TenantName,
		Address: model.Address{
			Street:     lesson.Tenant.TenantAddress.Street,
			PostalCode: lesson.Tenant.TenantAddress.PostalCode,
			City:       lesson.Tenant.TenantAddress.City,
			Country:    lesson.Tenant.TenantAddress.Country,
		},
		Link: fmt.Sprintf("https://app.playtomic.io/tenant/%s", lesson.Tenant.TenantID),
	}

	// Extract level info from lesson description if available
	if lesson.LevelDescription != "" {
		// Try to parse min and max levels from the level description
		// Format is typically "X.X - Y.Y"
		var minLvl, maxLvl float64
		n, _ := fmt.Sscanf(lesson.LevelDescription, "%f - %f", &minLvl, &maxLvl)
		if n == 2 {
			activity.MinLevel = minLvl
			activity.MaxLevel = maxLvl
		}
	}

	// Handle registered players
	minLevel := math.MaxFloat64
	maxLevel := 0.0

	for _, player := range lesson.RegisteredPlayers {
		// Create our player model
		p := model.Player{
			ID:    player.UserID,
			Name:  player.FullName,
			Level: player.LevelValue,
			Team:  nil, // Lessons don't have teams
			Link:  fmt.Sprintf("https://app.playtomic.io/profile/user/%s", player.UserID),
		}

		activity.RegisteredPlayers = append(activity.RegisteredPlayers, p)

		// Track min/max levels
		if player.LevelValue < minLevel {
			minLevel = player.LevelValue
		}
		if player.LevelValue > maxLevel {
			maxLevel = player.LevelValue
		}
	}

	// If we didn't successfully parse the level range from description,
	// use the min/max from registered players
	if activity.MinLevel == 0 && activity.MaxLevel == 0 && len(activity.RegisteredPlayers) > 0 {
		activity.MinLevel = minLevel
		activity.MaxLevel = maxLevel
	}

	// If we still don't have levels, set defaults
	if activity.MinLevel == 0 && activity.MaxLevel == 0 {
		activity.MinLevel = 1.0
		activity.MaxLevel = 5.0
	}

	// Link to the tournament
	activity.Link = fmt.Sprintf("https://app.playtomic.io/training/%s", lesson.TournamentID)

	return activity, nil
}

// ExternalLessonsToActivities transforms a slice of external lessons to our domain Activity models
func ExternalLessonsToActivities(lessons []models.Lesson) ([]model.Activity, error) {
	activities := make([]model.Activity, 0, len(lessons))

	for _, lesson := range lessons {
		activity, err := ExternalLessonToActivity(lesson)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}
