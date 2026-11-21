// Package characters owns player and non-player characters and their per-system sheets.
package characters

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

// Character kinds.
const (
	KindPC  = "pc"
	KindNPC = "npc"
)

var (
	// ErrNotFound means the character does not exist or is not visible to the caller.
	ErrNotFound = errors.New("character not found")
	// ErrForbidden means the caller may not change this character.
	ErrForbidden = errors.New("forbidden")
	// ErrInvalidInput means a request field failed validation.
	ErrInvalidInput = errors.New("invalid input")
	// ErrNoSheetSupport means the campaign system has no sheet module.
	ErrNoSheetSupport = errors.New("system has no sheets")
)

// RejectedError carries the issues that stopped a sheet from being saved.
type RejectedError struct{ Issues []systems.Issue }

func (e *RejectedError) Error() string { return "sheet rejected" }

// Character is a stored character; Sheet is opaque to this package.
type Character struct {
	ID          string          `json:"id"`
	CampaignID  string          `json:"campaignId"`
	OwnerID     *string         `json:"ownerId"`
	OwnerName   string          `json:"ownerName"`
	Kind        string          `json:"kind"`
	Name        string          `json:"name"`
	PortraitURL *string         `json:"portraitUrl"`
	Sheet       json.RawMessage `json:"sheet"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// View is a character as sent to clients, with computed values and validation findings.
type View struct {
	Character
	Derived json.RawMessage `json:"derived,omitempty"`
	Issues  []systems.Issue `json:"issues"`
}
