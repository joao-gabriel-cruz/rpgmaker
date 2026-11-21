package actions

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/items"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// Scenes is the part of scenes.Service used here.
type Scenes interface {
	List(ctx context.Context, campaignID string) ([]scenes.Summary, error)
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
}

// Tokens is the part of the tokens store used here.
type Tokens interface {
	List(ctx context.Context, campaignID, sceneID string) ([]tokens.Token, error)
}

// Characters is the part of characters.Service used here.
type Characters interface {
	Get(ctx context.Context, m campaigns.Member, id string) (characters.View, error)
	EditSheet(ctx context.Context, campaignID, id string, fn func(json.RawMessage) (json.RawMessage, error)) (characters.View, error)
}

// Rolls is the part of rolls.Service used here.
type Rolls interface {
	RequestEntry(ctx context.Context, campaignID string, sender realtime.Sender, in rolls.RequestInput) (rolls.Entry, error)
	Final(ctx context.Context, campaignID, rollID string) (rolls.Final, error)
}

// Loot is the part of items.Service that reveals hidden loot.
type Loot interface {
	TryFind(ctx context.Context, campaignID, characterID, sceneID string, target items.Target, successes int) ([]items.Found, error)
}

// Effects is the part of effects.Service used to suggest consequences to the Mestre.
type Effects interface {
	Suggest(ctx context.Context, in effects.Suggestion) (effects.Suggestion, error)
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the catalog and the action runs waiting for their roll to be finalized.
type Service struct {
	store   Store
	scenes  Scenes
	tokens  Tokens
	chars   Characters
	rolls   Rolls
	loot    Loot
	effects Effects
	events  Events
	pick    func(n int) int

	mu   sync.Mutex
	runs map[string]*run
}

// NewService builds the service; the caller must register Finalized with rolls.OnFinalized.
func NewService(st Store, sc Scenes, tk Tokens, ch Characters, r Rolls, lt Loot, ef Effects, ev Events) *Service {
	return &Service{store: st, scenes: sc, tokens: tk, chars: ch, rolls: r, loot: lt, effects: ef, events: ev,
		pick: cryptoPick, runs: map[string]*run{}}
}

func cryptoPick(n int) int {
	v, err := crand.Int(crand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(v.Int64())
}

// run is an action waiting for its roll; it lives in memory only, so a restart drops pending runs.
type run struct {
	id      string
	rollID  string
	act     Action
	char    characters.View
	actor   tokens.Token
	target  Target
	victim  *tokens.Token
	sceneID string
	safe    bool
	variant string
}

func (s *Service) remember(r *run) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.runs) >= maxRuns {
		s.runs = map[string]*run{}
	}
	s.runs[r.rollID] = r
}

func (s *Service) take(rollID string) *run {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.runs[rollID]
	delete(s.runs, rollID)
	return r
}

func newRunID() string {
	var b [8]byte
	_, _ = crand.Read(b[:])
	return hex.EncodeToString(b[:])
}
