package transformer

import (
	"fmt"
	"math"
	"time"

	"github.com/yourusername/padel-alert/internal/domain/model"
	"github.com/yourusername/padel-alert/pkg/playtomic"
)

const playtomicTimeFormat = "2006-01-02T15:04:05"

func ClassToActivity(class playtomic.Class) (model.Activity, error) {
	activity := model.Activity{
		ID:                class.AcademyClassID,
		Type:              "ACADEMY_CLASS",
		PlaytomicType:     class.Type,
		RegisteredPlayers: []model.Player{},
	}

	startDate, err := time.Parse(playtomicTimeFormat, class.StartDate)
	if err != nil {
		return activity, fmt.Errorf("parsing start date: %w", err)
	}
	activity.StartDate = startDate

	endDate, err := time.Parse(playtomicTimeFormat, class.EndDate)
	if err != nil {
		return activity, fmt.Errorf("parsing end date: %w", err)
	}
	activity.EndDate = endDate

	duration := endDate.Sub(startDate).Minutes()
	activity.Duration = int(duration)

	activity.Club = model.Club{
		ID:   class.Tenant.TenantID,
		Name: class.Tenant.TenantName,
		Address: model.Address{
			Street:     class.Tenant.Address.Street,
			PostalCode: class.Tenant.Address.PostalCode,
			City:       class.Tenant.Address.City,
			Country:    class.Tenant.Address.Country,
		},
		Link: fmt.Sprintf("https://app.playtomic.io/tenant/%s", class.Tenant.TenantID),
	}

	activity.Price = class.RegistrationInfo.BasePrice

	if class.CourseSummary != nil {
		activity.Name = class.CourseSummary.Name
		activity.Gender = class.CourseSummary.Gender
		activity.MinPlayers = class.CourseSummary.MinPlayers
		activity.MaxPlayers = class.CourseSummary.MaxPlayers

		registeredCount := len(class.RegistrationInfo.Registrations)
		activity.AvailablePlaces = activity.MaxPlayers - registeredCount
		if activity.AvailablePlaces < 0 {
			activity.AvailablePlaces = 0
		}
	} else {
		activity.Name = class.Resource.Name
		activity.Gender = "UNRESTRICTED"
		activity.MinPlayers = 1
		activity.MaxPlayers = len(class.RegistrationInfo.Registrations)
		activity.AvailablePlaces = 0
	}

	minLevel := math.MaxFloat64
	maxLevel := 0.0

	for _, registration := range class.RegistrationInfo.Registrations {
		player := model.Player{
			ID:    registration.Player.UserID,
			Name:  registration.Player.Name,
			Level: registration.Player.LevelValue,
			Team:  nil,
			Link:  fmt.Sprintf("https://app.playtomic.io/profile/user/%s", registration.Player.UserID),
		}

		activity.RegisteredPlayers = append(activity.RegisteredPlayers, player)

		if registration.Player.LevelValue < minLevel {
			minLevel = registration.Player.LevelValue
		}
		if registration.Player.LevelValue > maxLevel {
			maxLevel = registration.Player.LevelValue
		}
	}

	if len(activity.RegisteredPlayers) > 0 {
		activity.MinLevel = minLevel
		activity.MaxLevel = maxLevel
	} else {
		activity.MinLevel = 0.0
		activity.MaxLevel = 0.0
	}

	activity.Link = fmt.Sprintf("https://app.playtomic.io/lesson_class/%s",
		class.AcademyClassID)

	return activity, nil
}

func ClassesToActivities(classes []playtomic.Class) ([]model.Activity, error) {
	activities := make([]model.Activity, 0, len(classes))

	for _, class := range classes {
		activity, err := ClassToActivity(class)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func MatchToActivity(match playtomic.Match) (model.Activity, error) {
	// Set type based on match_type (COMPETITIVE or FRIENDLY)
	matchType := "MATCH_COMPETITIVE"
	if match.MatchType == "FRIENDLY" {
		matchType = "MATCH_FRIENDLY"
	}

	activity := model.Activity{
		ID:                match.MatchID,
		Type:              matchType,
		PlaytomicType:     match.MatchType,
		RegisteredPlayers: []model.Player{},
	}

	// Since matches don't have a name field, use format "Padel Match at [Club Name]"
	activity.Name = fmt.Sprintf("Padel Match at %s", match.Location)

	startDate, err := time.Parse(playtomicTimeFormat, match.StartDate)
	if err != nil {
		return activity, fmt.Errorf("parsing start date: %w", err)
	}
	activity.StartDate = startDate

	endDate, err := time.Parse(playtomicTimeFormat, match.EndDate)
	if err != nil {
		return activity, fmt.Errorf("parsing end date: %w", err)
	}
	activity.EndDate = endDate

	duration := endDate.Sub(startDate).Minutes()
	activity.Duration = int(duration)

	activity.Club = model.Club{
		ID:   match.Tenant.TenantID,
		Name: match.Tenant.TenantName,
		Address: model.Address{
			Street:     match.Tenant.Address.Street,
			PostalCode: match.Tenant.Address.PostalCode,
			City:       match.Tenant.Address.City,
			Country:    match.Tenant.Address.Country,
		},
		Link: fmt.Sprintf("https://app.playtomic.io/tenant/%s", match.Tenant.TenantID),
	}

	activity.Price = match.Price
	activity.Gender = match.Gender
	activity.MinLevel = match.MinLevel
	activity.MaxLevel = match.MaxLevel

	// Calculate min and max players by summing across teams
	totalMinPlayers := 0
	totalMaxPlayers := 0
	registeredCount := 0

	for _, team := range match.Teams {
		totalMinPlayers += team.MinPlayers
		totalMaxPlayers += team.MaxPlayers
		registeredCount += len(team.Players)

		// Add players from each team
		for _, p := range team.Players {
			// Map team ID "0" to "A" and "1" to "B"
			var teamName string
			switch team.TeamID {
			case "0":
				teamName = "A"
			case "1":
				teamName = "B"
			default:
				teamName = team.TeamID
			}

			player := model.Player{
				ID:    p.UserID,
				Name:  p.Name,
				Level: p.LevelValue,
				Team:  &teamName,
				Link:  fmt.Sprintf("https://app.playtomic.io/profile/user/%s", p.UserID),
			}
			activity.RegisteredPlayers = append(activity.RegisteredPlayers, player)
		}
	}

	activity.MinPlayers = totalMinPlayers
	activity.MaxPlayers = totalMaxPlayers
	activity.AvailablePlaces = totalMaxPlayers - registeredCount

	if activity.AvailablePlaces < 0 {
		activity.AvailablePlaces = 0
	}

	activity.Link = fmt.Sprintf("https://app.playtomic.io/padel-match/%s", match.MatchID)

	return activity, nil
}

func MatchesToActivities(matches []playtomic.Match) ([]model.Activity, error) {
	activities := make([]model.Activity, 0, len(matches))

	for _, match := range matches {
		activity, err := MatchToActivity(match)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}
