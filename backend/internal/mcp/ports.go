package mcp

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// Scenes is the slice of scenes.Service the MCP tools need.
type Scenes interface {
	List(ctx context.Context, campaignID string) ([]scenes.Summary, error)
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
	Paint(ctx context.Context, campaignID, sceneID, origin string, p maps.Patch) error
	EditObjects(ctx context.Context, campaignID, sceneID, origin string, p scenes.ObjectPatch) error
	Create(ctx context.Context, campaignID, name string, w, d int) (scenes.Scene, error)
	Activate(ctx context.Context, campaignID, id string) error
}

// Tokens is the slice of tokens.Service the MCP tools need.
type Tokens interface {
	List(ctx context.Context, m campaigns.Member, sceneID string) ([]tokens.Token, error)
	Move(ctx context.Context, snd realtime.Sender, campaignID, tokenID string, x, z int, force bool) error
	Place(ctx context.Context, snd realtime.Sender, campaignID, sceneID, charID string, x, z int) error
	Remove(ctx context.Context, snd realtime.Sender, campaignID, tokenID string) error
}

// Ambience is the slice of ambience.Service the MCP tools need.
type Ambience interface {
	Get(ctx context.Context, campaignID, sceneID string) (ambience.Ambience, error)
	Set(ctx context.Context, campaignID string, a ambience.Ambience) error
	SetCondition(ctx context.Context, snd realtime.Sender, campaignID, charID, cond string, active bool) error
}

// Characters is the slice of characters.Service the MCP tools need.
type Characters interface {
	List(ctx context.Context, m campaigns.Member) ([]characters.View, error)
	Create(ctx context.Context, m campaigns.Member, in characters.CreateInput) (characters.View, error)
	Save(ctx context.Context, m campaigns.Member, id string, sheet json.RawMessage) (characters.View, error)
}

// Table bundles the services the MCP tools act on.
type Table struct {
	Scenes     Scenes
	Tokens     Tokens
	Ambience   Ambience
	Characters Characters
}
