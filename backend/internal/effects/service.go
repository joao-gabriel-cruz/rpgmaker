package effects

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Scenes is the part of scenes.Service used here.
type Scenes interface {
	List(ctx context.Context, campaignID string) ([]scenes.Summary, error)
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
}

// Tokens is the part of the tokens store used here.
type Tokens interface {
	Get(ctx context.Context, campaignID, id string) (tokens.Token, error)
	List(ctx context.Context, campaignID, sceneID string) ([]tokens.Token, error)
}

// Characters is the part of characters.Service used here.
type Characters interface {
	Get(ctx context.Context, m campaigns.Member, id string) (characters.View, error)
	Save(ctx context.Context, m campaigns.Member, id string, sheet json.RawMessage) (characters.View, error)
}

// Rolls is the part of rolls.Service that owns the pending difficulty of a character.
type Rolls interface {
	State(ctx context.Context, m campaigns.Member, characterID string) (rolls.State, error)
	SetDifficulty(ctx context.Context, campaignID string, sender realtime.Sender, characterID string, d12 int) error
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the effect rules, triggers and the confirm/dismiss flow.
type Service struct {
	store  Store
	scenes Scenes
	tokens Tokens
	chars  Characters
	rolls  Rolls
	events Events
}

// NewService builds the service.
func NewService(st Store, sc Scenes, tk Tokens, ch Characters, r Rolls, ev Events) *Service {
	return &Service{store: st, scenes: sc, tokens: tk, chars: ch, rolls: r, events: ev}
}

// effective merges the campaign overrides over the defaults for one terrain or prefab.
func effective(over map[string][]Rule, kind, key string) []Rule {
	if r, ok := over[id(kind, key)]; ok {
		return r
	}
	return defaults[id(kind, key)]
}
