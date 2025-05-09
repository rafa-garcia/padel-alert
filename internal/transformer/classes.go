package transformer

import (
	"fmt"
	"math"

	"github.com/rafa-garcia/go-playtomic-api/models"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
)

// ExternalClassToActivity transforms an external Playtomic class to our domain Activity model
func ExternalClassToActivity(class models.Class) (model.Activity, error) {
	activity := model.Activity{
		ID:                class.AcademyClassID,
		Type:              "ACADEMY_CLASS",
		PlaytomicType:     class.Type,
		RegisteredPlayers: []model.Player{},
	}

	startDate, err := models.ParseTime(class.StartDate)
	if err != nil {
		return activity, fmt.Errorf("parsing start date: %w", err)
	}
	activity.StartDate = startDate

	endDate, err := models.ParseTime(class.EndDate)
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

// ExternalClassesToActivities transforms a slice of external classes to our domain Activity models
func ExternalClassesToActivities(classes []models.Class) ([]model.Activity, error) {
	activities := make([]model.Activity, 0, len(classes))

	for _, class := range classes {
		activity, err := ExternalClassToActivity(class)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}
