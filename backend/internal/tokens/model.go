// Package tokens places character tokens on scenes, validates their movement and relays pings.
package tokens

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// Errors returned by the service.
var (
	ErrNotFound = errors.New("tokens: not found")
	ErrExists   = errors.New("tokens: already on scene")
)

// Token is a character standee on a scene cell, with the character data clients need to draw it.
type Token struct {
	ID          string          `json:"id"`
	SceneID     string          `json:"sceneId"`
	CharacterID string          `json:"characterId"`
	X           int             `json:"x"`
	Z           int             `json:"z"`
	Props       json.RawMessage `json:"props"`
	Name        string          `json:"name"`
	PortraitURL *string         `json:"portraitUrl"`
	OwnerID     *string         `json:"ownerId"`
	// Appearance is the character's miniature description (sheet.appearance), null when unset.
	Appearance json.RawMessage `json:"appearance"`
	// AssimilationLevel comes from sheet.tugOfWar; null for sheets without it (NPCs).
	AssimilationLevel *int `json:"assimilationLevel"`
}

// Char is the character data the rules need.
type Char struct {
	ID      string
	Name    string
	OwnerID *string
}

// Repo is the persistence port.
type Repo interface {
	Place(ctx context.Context, sceneID, characterID string, x, z int) (Token, error)
	Get(ctx context.Context, campaignID, id string) (Token, error)
	List(ctx context.Context, campaignID, sceneID string) ([]Token, error)
	Move(ctx context.Context, id string, x, z int) error
	Remove(ctx context.Context, id string) error
	Character(ctx context.Context, campaignID, id string) (Char, error)
}

// SceneSource loads scenes (players only get the active one).
type SceneSource interface {
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
}

// Visibility filters the tokens a member may see (fog of war).
type Visibility interface {
	Filter(ctx context.Context, m campaigns.Member, sceneID string, list []Token) []Token
}
