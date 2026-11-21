// Package refuge owns the group's shared Refúgio: attributes, resources and a shared inventory.
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
	maxAttr       = 10
	maxText       = 5000
	maxName       = 100
)

var (
	// ErrCharNotFound means the character is missing or not in the campaign.
	ErrCharNotFound = errors.New("character not found")
	// ErrInvalid means a field failed validation.
	ErrInvalid = errors.New("invalid input")
)

// Details are the GM-editable fields of the Refúgio.
type Details struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Defense      int    `json:"defense"`
	Morale       int    `json:"morale"`
	Mobility     int    `json:"mobility"`
	Belligerence int    `json:"belligerence"`
	Water        int    `json:"water"`
	Food         int    `json:"food"`
	Wood         int    `json:"wood"`
	Notes        string `json:"notes"`
}

// Refuge is the campaign's Refúgio; items use the sheet's item format and are kept verbatim.
type Refuge struct {
	Details
	CampaignID string            `json:"campaignId"`
	Items      []json.RawMessage `json:"items"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

// CharState is the part of a character the move rules need.
type CharState struct {
	OwnerID *string
	Kind    string
	Sheet   json.RawMessage
}

func (d Details) validate() error {
	for _, v := range []int{d.Defense, d.Morale, d.Mobility, d.Belligerence} {
		if v < 0 || v > maxAttr {
			return ErrInvalid
		}
	}
	if d.Water < 0 || d.Food < 0 || d.Wood < 0 {
		return ErrInvalid
	}
	if len(d.Name) == 0 || len(d.Name) > maxName || len(d.Description) > maxText || len(d.Notes) > maxText {
		return ErrInvalid
	}
	return nil
}
