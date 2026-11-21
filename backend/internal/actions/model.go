// Package actions owns the action catalog (verbs a character performs on the table) and runs them:
// target -> roll -> animation for everyone -> consequence suggested to the Mestre.
package actions

import "errors"

// Categories, targets, outcomes and animations an action may use.
const (
	CatCombat      = "combat"
	CatExploration = "exploration"
	CatSocial      = "social"
	CatSurvival    = "survival"

	TargetNone   = "none"
	TargetCell   = "cell"
	TargetToken  = "token"
	TargetObject = "object"

	OutcomeNone   = "none"
	OutcomeDamage = "damage"
	OutcomeHeal   = "heal"
	OutcomeLoot   = "loot"
	OutcomeRest   = "rest"
	OutcomeMorale = "morale"
	OutcomeNotice = "notice"
)

// Events emitted to clients (names are a contract with the ambience/sound feature).
const (
	EventStarted  = "action.started"
	EventResolved = "action.resolved"
	EventCards    = "action.cards"
)

// Limits.
const (
	MaxName    = 40
	MaxRange   = 20
	MaxRolls   = 4
	maxRuns    = 500
	lootBonus  = 2
	safeRadius = 3
	eyeHeight  = 1.5
)

// ErrNotFound means the action does not exist in the campaign.
var ErrNotFound = errors.New("actions: not found")

// RollOption is one roll an action may use; the first is the default.
type RollOption struct {
	Instinct string `json:"instinct"`
	Aptitude string `json:"aptitude,omitempty"`
	Label    string `json:"label"`
}

// Action is a catalog entry.
type Action struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Icon      string       `json:"icon"`
	Category  string       `json:"category"`
	Target    string       `json:"target"`
	Rolls     []RollOption `json:"rolls"`
	Animation string       `json:"animation"`
	Outcome   string       `json:"outcome"`
	Range     int          `json:"range,omitempty"`
	Enabled   bool         `json:"enabled"`
	// Builtin marks ids that come from the system defaults (overridable, never deletable).
	Builtin bool `json:"builtin"`
}

// Target is where an action is aimed; X/Z are filled by the server for tokens and objects.
type Target struct {
	Kind     string `json:"kind"`
	X        *int   `json:"x,omitempty"`
	Z        *int   `json:"z,omitempty"`
	TokenID  string `json:"tokenId,omitempty"`
	ObjectID string `json:"objectId,omitempty"`
}

// Modifiers are the optional roll modifiers the player picked.
type Modifiers struct {
	Effort          bool   `json:"effort"`
	UseOccupation   bool   `json:"useOccupation"`
	UseMarkingEvent bool   `json:"useMarkingEvent"`
	SecondInstinct  string `json:"secondInstinct"`
}

// PerformInput is the payload of action.perform.
type PerformInput struct {
	CharacterID string    `json:"characterId"`
	ActionID    string    `json:"actionId"`
	Target      Target    `json:"target"`
	RollChoice  int       `json:"rollChoice"`
	Modifiers   Modifiers `json:"modifiers"`
}
