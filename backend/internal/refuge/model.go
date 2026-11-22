// Package refuge owns the group's shared Refúgio: attributes, population, reserves, resources,
// region, week counter and a shared inventory.
package refuge

import (
	"encoding/json"
	"errors"
	"time"
)

// Slot limits of a character sheet inventory.
const (
	BodySlots     = 3
	BackpackSlots = 6
	maxItems      = 200
	maxText       = 5000
	maxName       = 100
)

var (
	// ErrCharNotFound means the character is missing or not in the campaign.
	ErrCharNotFound = errors.New("character not found")
	// ErrInvalid means a field failed validation.
	ErrInvalid = errors.New("invalid input")
)

// Population is the group's size as a level (about 10 inhabitants each) bounded by Max.
type Population struct {
	Level int `json:"level"`
	Max   int `json:"max"`
}

// Region describes the surroundings: six 0-6 attributes plus abundance per natural resource.
type Region struct {
	Size             int               `json:"size"`
	Danger           int               `json:"danger"`
	Habitation       int               `json:"habitation"`
	NaturalResources int               `json:"naturalResources"`
	Contamination    int               `json:"contamination"`
	Travel           int               `json:"travel"`
	Abundance        map[string]string `json:"abundance"`
}

// Text holds the free-text fields of the Refúgio.
type Text struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
}

// Sheet is the numeric state a delivery may edit; stored in the refuges.state jsonb column.
type Sheet struct {
	Defense        int            `json:"defense"`
	Morale         int            `json:"morale"`
	Belligerence   int            `json:"belligerence"`
	Mobility       int            `json:"mobility"`
	Population     Population     `json:"population"`
	ReservesMax    int            `json:"reservesMax"`
	HasWaterSource bool           `json:"hasWaterSource"`
	Resources      map[string]int `json:"resources"`
	Region         Region         `json:"region"`
	Week           int            `json:"week"`
}

// Editable is what the Mestre can replace through PUT: text plus the Sheet.
type Editable struct {
	Text
	Sheet
}

// Refuge is the campaign's Refúgio. Crises and Residents are reserved for the next delivery
// (kept verbatim, empty for now); items use the sheet's item format.
type Refuge struct {
	CampaignID string `json:"campaignId"`
	Editable
	Buildings []Building        `json:"buildings"`
	Proposals []Proposal        `json:"proposals"`
	History   []WeekReport      `json:"history"`
	Crises    []Crisis          `json:"crises"`
	Residents []Resident        `json:"residents"`
	Items     []json.RawMessage `json:"items"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Derived   Derived           `json:"derived"`

	Expeditions    []Expedition   `json:"expeditions"`
	CrisisSettings CrisisSettings `json:"crisisSettings"`
	// Conditions are derived alerts (e.g. "disease"), never stored.
	Conditions []string `json:"conditions"`
}

// CharState is the part of a character the move rules need.
type CharState struct {
	OwnerID *string
	Kind    string
	Sheet   json.RawMessage
}
