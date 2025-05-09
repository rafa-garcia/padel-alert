package transformer

import (
	"fmt"

	"github.com/rafa-garcia/go-playtomic-api/models"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
)

// ExternalMatchToActivity transforms an external Playtomic match to our domain Activity model
func ExternalMatchToActivity(match models.Match) (model.Activity, error) {
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

	startDate, err := models.ParseTime(match.StartDate)
	if err != nil {
		return activity, fmt.Errorf("parsing start date: %w", err)
	}
	activity.StartDate = startDate

	endDate, err := models.ParseTime(match.EndDate)
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
			// Map team IDs "0"/"1" to "A"/"B"
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

// ExternalMatchesToActivities transforms a slice of external matches to our domain Activity models
func ExternalMatchesToActivities(matches []models.Match) ([]model.Activity, error) {
	activities := make([]model.Activity, 0, len(matches))

	for _, match := range matches {
		activity, err := ExternalMatchToActivity(match)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}
