// Package drawings persists free-hand strokes drawn over a scene and relays the ephemeral ruler.
package drawings

import (
	"context"
	"errors"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// ErrNotFound is returned when a drawing does not exist in the campaign.
var ErrNotFound = errors.New("drawings: not found")

// Limits enforced on every stroke.
const (
	MaxPoints = 500
	MinWidth  = 1
	MaxWidth  = 10
	maxCoord  = 1000.0
)

// Point is a world-space position on the ground plane (cell units).
type Point struct {
	X float64 `json:"x"`
	Z float64 `json:"z"`
}

// Drawing is one stroke; heights are derived from the terrain by the client.
type Drawing struct {
	ID        string    `json:"id"`
	SceneID   string    `json:"sceneId"`
	UserID    string    `json:"userId"`
	Color     string    `json:"color"`
	Width     int       `json:"width"`
	Points    []Point   `json:"points"`
	CreatedAt time.Time `json:"createdAt"`
}

// Repo is the persistence port.
type Repo interface {
	Add(ctx context.Context, d Drawing) (Drawing, error)
	Get(ctx context.Context, campaignID, id string) (Drawing, error)
	List(ctx context.Context, campaignID, sceneID string) ([]Drawing, error)
	Remove(ctx context.Context, id string) error
	Clear(ctx context.Context, sceneID string) error
}

// SceneSource loads scenes (players only get the active one).
type SceneSource interface {
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
}
