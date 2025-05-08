package model

import (
	"time"
)

// Activity represents a padel activity (academy class, tournament or match)
type Activity struct {
	ID                string    `json:"id"`
	Type              string    `json:"type"`
	PlaytomicType     string    `json:"playtomic_type"`
	Name              string    `json:"name"`
	Club              Club      `json:"club"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	Duration          int       `json:"duration"` // Duration in minutes
	MinPlayers        int       `json:"min_players"`
	MaxPlayers        int       `json:"max_players"`
	MinLevel          float64   `json:"min_level"`
	MaxLevel          float64   `json:"max_level"`
	Price             string    `json:"price"`
	Gender            string    `json:"gender"`
	AvailablePlaces   int       `json:"available_places"`
	RegisteredPlayers []Player  `json:"registered_players"`
	Link              string    `json:"link"`
}

// Club represents a padel club
type Club struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Address Address `json:"address"`
	Link    string  `json:"link"`
}

// Address represents a physical address
type Address struct {
	Street     string `json:"street"`
	PostalCode string `json:"postal_code"`
	City       string `json:"city"`
	Country    string `json:"country"`
}

// Player represents a registered player
type Player struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Level float64 `json:"level"`
	Team  *string `json:"team,omitempty"`
	Link  string  `json:"link"`
}
