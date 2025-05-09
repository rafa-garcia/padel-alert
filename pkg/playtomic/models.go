package playtomic

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const playtomicTimeFormat = "2006-01-02T15:04:05"

// Match represents a match from the Playtomic API
type Match struct {
	MatchID                       string             `json:"match_id"`
	ReservationID                 *string            `json:"reservation_id"`
	RecurringMatchConfigurationID *string            `json:"recurring_match_configuration_id"`
	Location                      string             `json:"location"`
	SportID                       string             `json:"sport_id"`
	Teams                         []Team             `json:"teams"`
	MinPlayersPerTeam             int                `json:"min_players_per_team"`
	MaxPlayersPerTeam             int                `json:"max_players_per_team"`
	OwnerID                       *string            `json:"owner_id"`
	Status                        string             `json:"status"`
	GameStatus                    string             `json:"game_status"`
	StartDate                     string             `json:"start_date"`
	EndDate                       string             `json:"end_date"`
	Tenant                        Tenant             `json:"tenant"`
	LocationInfo                  LocationInfo       `json:"location_info"`
	MatchType                     string             `json:"match_type"`
	MatchOrganization             string             `json:"match_organization"`
	CompetitionMode               string             `json:"competition_mode"`
	Gender                        string             `json:"gender"`
	MaxLevel                      float64            `json:"max_level"`
	MinLevel                      float64            `json:"min_level"`
	Price                         string             `json:"price"`
	PaymentRequired               bool               `json:"payment_required"`
	ResourceProperties            ResourceProperties `json:"resource_properties"`
	RegistrationInfo              RegistrationInfo   `json:"registration_info"`
	MatchOrigin                   string             `json:"match_origin"`
	RegistrationType              string             `json:"registration_type"`
	RegistrationStatus            string             `json:"registration_status"`
	IsPremium                     bool               `json:"is_premium"`
	IsBooked                      bool               `json:"is_booked"`
	CreatedAt                     string             `json:"created_at"`
	Visibility                    string             `json:"visibility"`
}

// Team represents a team in a match
type Team struct {
	TeamID     string   `json:"team_id"`
	Players    []Player `json:"players"`
	MinPlayers int      `json:"min_players"`
	MaxPlayers int      `json:"max_players"`
	TeamResult *string  `json:"team_result"`
}

// LocationInfo represents information about the location of a match
type LocationInfo struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Address Address  `json:"address"`
	Images  []string `json:"images"`
}

// ResourceProperties contains properties of a resource used for a match
type ResourceProperties struct {
	ResourceType    string `json:"resource_type"`
	ResourceSize    string `json:"resource_size"`
	ResourceFeature string `json:"resource_feature"`
}

// ParsePlaytomicTime parses a time string in Playtomic's format
func ParsePlaytomicTime(timeStr string) (time.Time, error) {
	return time.Parse(playtomicTimeFormat, timeStr)
}

// CourseSummary represents summary information about a course
type CourseSummary struct {
	CourseID   string `json:"course_id"`
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Visibility string `json:"visibility"`
	MinPlayers int    `json:"min_players"`
	MaxPlayers int    `json:"max_players"`
}

// Class represents a class from the Playtomic API
type Class struct {
	Type             string           `json:"type"`
	AcademyClassID   string           `json:"academy_class_id"`
	SportID          string           `json:"sport_id"`
	Tenant           Tenant           `json:"tenant"`
	Resource         Resource         `json:"resource"`
	StartDate        string           `json:"start_date"`
	EndDate          string           `json:"end_date"`
	Coaches          []Coach          `json:"coaches"`
	RegistrationInfo RegistrationInfo `json:"registration_info"`
	CourseSummary    *CourseSummary   `json:"course_summary,omitempty"`
	AccessCode       *string          `json:"access_code"`
	Origin           string           `json:"origin"`
	IsCanceled       bool             `json:"is_canceled"`
	PrivateNotes     *string          `json:"private_notes"`
	PublicNotes      string           `json:"public_notes"`
	Status           string           `json:"status"`
	PaymentStatus    string           `json:"payment_status"`
}

// Tenant represents a club/venue in the Playtomic API
type Tenant struct {
	TenantID        string                 `json:"tenant_id"`
	TenantName      string                 `json:"tenant_name"`
	Address         Address                `json:"address"`
	Images          []string               `json:"images"`
	Properties      map[string]interface{} `json:"properties"`
	PlaytomicStatus string                 `json:"playtomic_status"`
}

// Address represents a physical address
type Address struct {
	Street                string     `json:"street"`
	PostalCode            string     `json:"postal_code"`
	City                  string     `json:"city"`
	SubAdministrativeArea string     `json:"sub_administrative_area"`
	AdministrativeArea    string     `json:"administrative_area"`
	Country               string     `json:"country"`
	CountryCode           string     `json:"country_code"`
	Coordinate            Coordinate `json:"coordinate"`
	Timezone              string     `json:"timezone"`
}

// Coordinate represents a geographical coordinate
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Resource represents a court or other resource
type Resource struct {
	ID         string                 `json:"id"`
	LockID     string                 `json:"lock_id"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

// Coach represents a coach
type Coach struct {
	UserID                 string  `json:"user_id"`
	LockID                 string  `json:"lock_id"`
	Name                   string  `json:"name"`
	Picture                string  `json:"picture"`
	Email                  *string `json:"email"`
	Gender                 *string `json:"gender"`
	LevelValue             float64 `json:"level_value"`
	LevelValueConfidence   float64 `json:"level_value_confidence"`
	Phone                  *string `json:"phone"`
	CommunicationsLanguage string  `json:"communications_language"`
	IsPremium              bool    `json:"is_premium"`
}

// RegistrationInfo represents registration information
type RegistrationInfo struct {
	PaymentType          string         `json:"payment_type"`
	NumberOfPlayers      int            `json:"number_of_players"`
	BasePrice            string         `json:"base_price"`
	IsManualPrice        bool           `json:"is_manual_price"`
	Registrations        []Registration `json:"registrations"`
	OnlinePaymentAllowed bool           `json:"online_payment_allowed"`
}

// Registration represents a registration
type Registration struct {
	ClassRegistrationID      string      `json:"class_registration_id"`
	Player                   Player      `json:"player"`
	Price                    string      `json:"price"`
	RegistrationDate         string      `json:"registration_date"`
	Payment                  Payment     `json:"payment"`
	CustomPriceConfiguration interface{} `json:"custom_price_configuration"`
	CustomPrice              string      `json:"custom_price"`
	IsManualPrice            bool        `json:"is_manual_price"`
	CourseBillID             *string     `json:"course_bill_id"`
	CategoryName             *string     `json:"category_name"`
}

// Player represents a player in a registration
type Player struct {
	UserID                 string  `json:"user_id"`
	GuestID                *string `json:"guest_id"`
	Name                   string  `json:"name"`
	Picture                string  `json:"picture"`
	Email                  *string `json:"email"`
	Gender                 *string `json:"gender"`
	LevelValue             float64 `json:"level_value"`
	LevelValueConfidence   float64 `json:"level_value_confidence"`
	Phone                  *string `json:"phone"`
	CommunicationsLanguage string  `json:"communications_language"`
	IsPremium              bool    `json:"is_premium"`
	FamilyMemberID         *string `json:"family_member_id"`
}

// Payment represents payment information
type Payment struct {
	PaymentID               string  `json:"payment_id"`
	PaymentMethodID         string  `json:"payment_method_id"`
	PaymentMethodType       string  `json:"payment_method_type"`
	OnsitePaymentMethodType *string `json:"onsite_payment_method_type"`
	B2bBillingType          string  `json:"b2b_billing_type"`
	UserVat                 float64 `json:"user_vat"`
	TenantVat               float64 `json:"tenant_vat"`
	CommissionModel         string  `json:"commission_model"`
	RefundID                *string `json:"refund_id"`
	PaymentPrice            string  `json:"payment_price"`
	PaymentReference        *string `json:"payment_reference"`
	PayerID                 string  `json:"payer_id"`
	PaymentDate             string  `json:"payment_date"`
}

// SearchClassesParams defines parameters for searching classes
type SearchClassesParams struct {
	Sort             string      `url:"sort,omitempty"`
	Status           string      `url:"status,omitempty"`
	Type             string      `url:"type,omitempty"`
	TenantIDs        []string    `url:"-"` // Will be joined with comma and set as tenant_id
	IncludeSummary   bool        `url:"include_summary,omitempty"`
	Size             int         `url:"size,omitempty"`
	Page             int         `url:"page,omitempty"`
	CourseVisibility string      `url:"course_visibility,omitempty"`
	FromStartDate    string      `url:"-"` // Will be formatted and set as from_start_date
	Coordinate       *Coordinate `url:"-"` // Will be formatted and set as coordinate
	Radius           int         `url:"radius,omitempty"`
}

// ToURLValues converts SearchClassesParams to url.Values
func (p *SearchClassesParams) ToURLValues() url.Values {
	values := url.Values{}

	if s := strings.TrimSpace(p.Sort); s != "" {
		values.Set("sort", s)
	}

	if s := strings.TrimSpace(p.Status); s != "" {
		values.Set("status", s)
	}

	if t := strings.TrimSpace(p.Type); t != "" {
		values.Set("type", t)
	}

	if len(p.TenantIDs) > 0 {
		values.Set("tenant_id", strings.Join(p.TenantIDs, ","))
	}

	if p.IncludeSummary {
		values.Set("include_summary", "true")
	}

	if p.Size > 0 {
		values.Set("size", fmt.Sprintf("%d", p.Size))
	}

	values.Set("page", fmt.Sprintf("%d", p.Page))

	if cv := strings.TrimSpace(p.CourseVisibility); cv != "" {
		values.Set("course_visibility", cv)
	}

	// Add from_start_date if provided
	if p.FromStartDate != "" {
		values.Set("from_start_date", p.FromStartDate)
	}

	// Add coordinate if provided (and tenant_id is not provided)
	if p.Coordinate != nil && len(p.TenantIDs) == 0 {
		values.Set("coordinate", fmt.Sprintf("%f,%f", p.Coordinate.Lat, p.Coordinate.Lon))

		// Only set radius if coordinate is provided
		if p.Radius > 0 {
			values.Set("radius", fmt.Sprintf("%d", p.Radius))
		}
	}

	return values
}

// SearchMatchesParams defines parameters for searching matches
type SearchMatchesParams struct {
	Sort          string   `url:"-"` // Will be joined with comma and set as sort
	HasPlayers    bool     `url:"-"` // Will be set as has_players
	SportID       string   `url:"-"` // Will be set as sport_id
	TenantIDs     []string `url:"-"` // Will be joined with comma and set as tenant_id
	Visibility    string   `url:"-"` // Will be set as visibility
	FromStartDate string   `url:"-"` // Will be formatted and set as from_start_date
	Size          int      `url:"-"` // Will be set as size
	Page          int      `url:"-"` // Will be set as page
}

// ToURLValues converts SearchMatchesParams to url.Values
func (p *SearchMatchesParams) ToURLValues() url.Values {
	values := url.Values{}

	if s := strings.TrimSpace(p.Sort); s != "" {
		values.Set("sort", s)
	}

	if p.HasPlayers {
		values.Set("has_players", "true")
	}

	if s := strings.TrimSpace(p.SportID); s != "" {
		values.Set("sport_id", s)
	}

	if len(p.TenantIDs) > 0 {
		values.Set("tenant_id", strings.Join(p.TenantIDs, ","))
	}

	if v := strings.TrimSpace(p.Visibility); v != "" {
		values.Set("visibility", v)
	}

	// Add from_start_date if provided
	if p.FromStartDate != "" {
		values.Set("from_start_date", p.FromStartDate)
	}

	if p.Size > 0 {
		values.Set("size", fmt.Sprintf("%d", p.Size))
	}

	if p.Page > 0 {
		values.Set("page", fmt.Sprintf("%d", p.Page))
	}

	return values
}
