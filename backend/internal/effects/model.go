// Package effects turns terrain and prefab rules into suggestions the Mestre confirms or dismisses.
package effects

import (
	"errors"
	"time"
)

// Rule triggers.
const (
	TriggerEnter     = "on_enter"
	TriggerTurnStart = "on_turn_start"
)

// Rule effects.
const (
	EffectDamage     = "damage"
	EffectDifficulty = "difficulty"
	EffectNotice     = "notice"
)

// Rule sources.
const (
	KindTerrain = "terrain"
	KindPrefab  = "prefab"
)

// Limits.
const (
	MaxRulesPerKey = 10
	MaxMessageLen  = 200
	MaxDamage      = 99
	MaxDifficulty  = 3
)

// Statuses of a suggestion.
const (
	StatusPending   = "pending"
	StatusApplied   = "applied"
	StatusDismissed = "dismissed"
)

// Events emitted to clients.
const (
	EventSuggested = "effect.suggested"
	EventResolved  = "effect.resolved"
	EventApplied   = "effect.applied"
)

// ErrNotFound means the suggestion does not exist, is not pending or belongs to another campaign.
var ErrNotFound = errors.New("effects: suggestion not found")

// Rule is one effect a terrain type or prefab causes.
type Rule struct {
	Trigger string `json:"trigger"`
	Effect  string `json:"effect"`
	Amount  int    `json:"amount"`
	Message string `json:"message"`
}

// Suggestion is a triggered effect waiting for the Mestre.
type Suggestion struct {
	ID            string    `json:"id"`
	CampaignID    string    `json:"campaignId"`
	CharacterID   string    `json:"characterId"`
	CharacterName string    `json:"characterName"`
	SceneID       string    `json:"sceneId"`
	X             int       `json:"x"`
	Z             int       `json:"z"`
	Source        string    `json:"source"`
	Trigger       string    `json:"trigger"`
	Effect        string    `json:"effect"`
	Amount        int       `json:"amount"`
	Message       string    `json:"message"`
	CreatedAt     time.Time `json:"createdAt"`
}

// RuleSet is the effective rules of one terrain type or prefab, as listed to the Mestre.
type RuleSet struct {
	Kind           string `json:"kind"`
	Key            string `json:"key"`
	Label          string `json:"label"`
	Rules          []Rule `json:"rules"`
	IsDefault      bool   `json:"isDefault"`
	BlocksMovement bool   `json:"blocksMovement"`
	BlocksVision   bool   `json:"blocksVision"`
}
