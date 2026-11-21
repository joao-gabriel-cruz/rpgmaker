// Package conflict holds the Partitura de Conflito: scene conditions, objectives, threats and activations.
package conflict

import (
	"errors"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
)

// Objective kinds and scopes.
const (
	KindMain        = "main"
	KindSecondary   = "secondary"
	ScopeCollective = "collective"
	ScopeIndividual = "individual"
	SourcePlayers   = "players"
	SourceThreat    = "threat:"
)

// Events broadcast to the table.
const (
	EventUpdated      = "conflict.updated"
	EventObjectiveHit = "conflict.objective.completed"
	EventThreatRolled = "conflict.threat.rolled"
	EventActivation   = "conflict.activation.triggered"
)

var (
	// ErrAlreadyAssigned means the roll already counted toward an objective.
	ErrAlreadyAssigned = errors.New("roll already assigned")
	// ErrNone means the campaign has no active conflict.
	ErrNone = errors.New("no active conflict")
)

// Objective is a goal with accumulated Successes.
type Objective struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Kind        string  `json:"kind"`
	Scope       string  `json:"scope"`
	CharacterID *string `json:"characterId,omitempty"`
	Target      int     `json:"target"`
	Progress    int     `json:"progress"`
}

// Threat has its own dice pool; Pool and Notes are hidden from players.
type Threat struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Pool           *rolls.ThreatPool `json:"pool,omitempty"`
	NpcCharacterID *string           `json:"npcCharacterId,omitempty"`
	Notes          string            `json:"notes,omitempty"`
}

// Activation fires when Count (symbols accumulated from Source) reaches Threshold.
type Activation struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	Threshold int    `json:"threshold"`
	Text      string `json:"text"`
	Source    string `json:"source"`
	Count     int    `json:"count"`
}

// Doc is the editable body stored as jsonb.
type Doc struct {
	Conditions  []string     `json:"conditions"`
	Objectives  []Objective  `json:"objectives"`
	Threats     []Threat     `json:"threats"`
	Activations []Activation `json:"activations,omitempty"`
}

// Conflict is the active score sheet of a campaign.
type Conflict struct {
	ID         string    `json:"id"`
	CampaignID string    `json:"-"`
	SceneID    *string   `json:"sceneId"`
	Title      string    `json:"title"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Doc
}

// ForPlayers hides threat pools, notes and activations.
func (c Conflict) ForPlayers() Conflict {
	out := c
	out.Activations = nil
	out.Threats = make([]Threat, len(c.Threats))
	for i, t := range c.Threats {
		out.Threats[i] = Threat{ID: t.ID, Name: t.Name, NpcCharacterID: t.NpcCharacterID}
	}
	return out
}
