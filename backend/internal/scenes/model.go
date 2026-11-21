// Package scenes persists campaign scenes (terrain grid + objects) and wires them to REST and realtime.
package scenes

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// Allowed scene side length, in cells.
const (
	MinSize = 16
	MaxSize = 64
)

// Limits protecting the server from oversized scenes and requests.
const (
	maxNameLen = 80
	maxObjects = 2000
)

// Errors returned by the service.
var (
	ErrNotFound     = errors.New("scenes: not found")
	ErrInvalidName  = errors.New("scenes: invalid name")
	ErrInvalidSize  = errors.New("scenes: invalid size")
	ErrInvalidPatch = errors.New("scenes: invalid patch")

	ErrTooManySnapshots = errors.New("scenes: too many snapshots")
)

// Scene is a full scene as returned by GET (grid + objects in one payload).
type Scene struct {
	ID         string          `json:"id"`
	CampaignID string          `json:"campaignId"`
	Name       string          `json:"name"`
	Width      int             `json:"width"`
	Depth      int             `json:"depth"`
	Grid       maps.Serialized `json:"grid"`
	Objects    []maps.Object   `json:"objects"`
	IsActive   bool            `json:"isActive"`
	Draft      *Draft          `json:"draft,omitempty"`
	Settings   json.RawMessage `json:"settings"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// Draft is the Mestre's hidden working copy of a scene; players never receive it.
type Draft struct {
	Grid    maps.Serialized `json:"grid"`
	Objects []maps.Object   `json:"objects"`
}

// Snapshot is a named saved state of a scene (terrain + objects).
type Snapshot struct {
	ID        string    `json:"id"`
	SceneID   string    `json:"sceneId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// SnapshotData is a snapshot with its payload.
type SnapshotData struct {
	Snapshot
	Grid    string
	Objects []maps.Object
}

// Summary is a scene without its heavy data, for listings.
type Summary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Width    int    `json:"width"`
	Depth    int    `json:"depth"`
	IsActive bool   `json:"isActive"`
}

// ObjectPatch is a batch of object changes (same shape as the frontend's).
type ObjectPatch struct {
	Add    []maps.Object `json:"add"`
	Remove []string      `json:"remove"`
	Update []maps.Object `json:"update"`
}

// Repo is the persistence port of the service.
type Repo interface {
	Create(ctx context.Context, campaignID, name string, w, d int, grid string) (Scene, error)
	List(ctx context.Context, campaignID string) ([]Summary, error)
	Get(ctx context.Context, campaignID, id string) (Scene, error)
	Rename(ctx context.Context, campaignID, id, name string) error
	Delete(ctx context.Context, campaignID, id string) error
	Activate(ctx context.Context, campaignID, id string) error
	SaveGrid(ctx context.Context, id, grid string) error
	SaveObjects(ctx context.Context, id string, objects []maps.Object) error
	// SaveDraft stores the draft (nil clears it).
	SaveDraft(ctx context.Context, id string, d *Draft) error
	// Replace overwrites the real grid+objects and clears the draft atomically.
	Replace(ctx context.Context, id, grid string, objects []maps.Object) error
	CreateSnapshot(ctx context.Context, sceneID, name, grid string, objects []maps.Object) (Snapshot, error)
	ListSnapshots(ctx context.Context, sceneID string) ([]Snapshot, error)
	GetSnapshot(ctx context.Context, sceneID, id string) (SnapshotData, error)
	DeleteSnapshot(ctx context.Context, sceneID, id string) error
}
