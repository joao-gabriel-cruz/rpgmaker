// Package vision holds per-scene fog of war: manual reveals, dynamic line of sight per player and the
// filtering of token events so players never learn about what they cannot see.
package vision

import (
	"context"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// Fog modes.
const (
	ModeOff     = "off"
	ModeManual  = "manual"
	ModeDynamic = "dynamic"
)

// Defaults and limits.
const (
	DefaultRadius = 12
	MaxRadius     = 40
	maxBrushCells = 64 * 64
	eyeHeight     = 1.5
)

// Events emitted to clients.
const (
	EventVision = "vision.updated"
	EventFog    = "fog.updated"
)

// ErrNotFound is returned when the scene does not exist (or is not visible to the member).
var ErrNotFound = errors.New("vision: not found")

// Fog is the stored fog configuration of a scene.
type Fog struct {
	Mode     string
	Radius   int
	Revealed string
}

// Pos is a token position with the owner of its character.
type Pos struct {
	ID      string
	X, Z    int
	OwnerID *string
}

// Repo is the persistence port.
type Repo interface {
	// Fog returns the stored fog, or the defaults (off, radius 12, nothing revealed) when none exists.
	Fog(ctx context.Context, sceneID string) (Fog, error)
	SaveConfig(ctx context.Context, sceneID, mode string, radius int) error
	SaveRevealed(ctx context.Context, sceneID, revealed string) error
	Positions(ctx context.Context, sceneID string) ([]Pos, error)
}

// SceneSource loads scenes.
type SceneSource interface {
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
}

// Hub is the part of the realtime hub the service needs.
type Hub interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
	Online(campaignID string) []realtime.Sender
}

// View is what a member gets from GET: the Mestre sees the config and manual reveals, a player only their visibility.
type View struct {
	SceneID      string `json:"sceneId"`
	Mode         string `json:"mode"`
	VisionRadius int    `json:"visionRadius,omitempty"`
	Revealed     string `json:"revealed,omitempty"`
	Visible      string `json:"visible,omitempty"`
}
