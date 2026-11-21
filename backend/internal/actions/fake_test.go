package actions

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/items"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

const (
	camp     = "00000000-0000-0000-0000-00000000000c"
	sceneID  = "00000000-0000-0000-0000-0000000000a1"
	charA    = "00000000-0000-0000-0000-0000000000b1"
	charNPC  = "00000000-0000-0000-0000-0000000000b2"
	ownerA   = "u-ana"
	otherUsr = "u-bea"
)

var ctx = context.Background()

type fakeStore struct{ list []Action }

func (f *fakeStore) List(context.Context, string) ([]Action, error) { return f.list, nil }
func (f *fakeStore) Save(_ context.Context, _ string, a Action) error {
	for i := range f.list {
		if f.list[i].ID == a.ID {
			f.list[i] = a
			return nil
		}
	}
	f.list = append(f.list, a)
	return nil
}
func (f *fakeStore) Delete(_ context.Context, _, id string) error {
	for i := range f.list {
		if f.list[i].ID == id {
			f.list = append(f.list[:i], f.list[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

type fakeScenes struct{ active bool }

func (f fakeScenes) List(context.Context, string) ([]scenes.Summary, error) {
	return []scenes.Summary{{ID: sceneID, IsActive: f.active}}, nil
}

func (fakeScenes) Get(context.Context, string, string, string) (scenes.Scene, error) {
	g, _ := maps.NewGrid(12, 12)
	return scenes.Scene{ID: sceneID, Width: 12, Depth: 12, Grid: g.Serialize(), Objects: []maps.Object{
		{ID: "crate1", Prefab: "crate", X: 5, Z: 5}, {ID: "fire1", Prefab: "campfire", X: 2, Z: 2}}}, nil
}

type fakeTokens struct{ list []tokens.Token }

func (f fakeTokens) List(context.Context, string, string) ([]tokens.Token, error) { return f.list, nil }

type fakeChars struct{ sheet json.RawMessage }

func (f *fakeChars) Get(_ context.Context, _ campaigns.Member, id string) (characters.View, error) {
	if id != charA {
		return characters.View{}, characters.ErrNotFound
	}
	owner := ownerA
	return characters.View{Character: characters.Character{ID: id, CampaignID: camp, Name: "Ana", OwnerID: &owner, Sheet: f.sheet}}, nil
}

func (f *fakeChars) EditSheet(_ context.Context, _, _ string, fn func(json.RawMessage) (json.RawMessage, error)) (characters.View, error) {
	out, err := fn(f.sheet)
	if err == nil {
		f.sheet = out
	}
	return characters.View{}, err
}

type fakeRolls struct {
	got  []rolls.RequestInput
	syms assimilacao.Symbols
}

func (f *fakeRolls) RequestEntry(_ context.Context, _ string, _ realtime.Sender, in rolls.RequestInput) (rolls.Entry, error) {
	f.got = append(f.got, in)
	return rolls.Entry{RollID: "roll-1"}, nil
}

func (f *fakeRolls) Final(context.Context, string, string) (rolls.Final, error) {
	return rolls.Final{Symbols: f.syms}, nil
}

type fakeLoot struct {
	target    items.Target
	successes int
	found     int
}

func (f *fakeLoot) TryFind(_ context.Context, _, _, _ string, t items.Target, n int) ([]items.Found, error) {
	f.target, f.successes = t, n
	return make([]items.Found, f.found), nil
}

type fakeEffects struct{ got []effects.Suggestion }

func (f *fakeEffects) Suggest(_ context.Context, in effects.Suggestion) (effects.Suggestion, error) {
	f.got = append(f.got, in)
	return in, nil
}

type sent struct {
	typ     string
	to      realtime.Filter
	payload any
}

type fakeEvents struct{ list []sent }

func (f *fakeEvents) Broadcast(_ string, to realtime.Filter, typ string, p any) {
	f.list = append(f.list, sent{typ, to, p})
}

func (f *fakeEvents) last(typ string) (sent, bool) {
	for i := len(f.list) - 1; i >= 0; i-- {
		if f.list[i].typ == typ {
			return f.list[i], true
		}
	}
	return sent{}, false
}

type rig struct {
	svc   *Service
	store *fakeStore
	chars *fakeChars
	rolls *fakeRolls
	loot  *fakeLoot
	ef    *fakeEffects
	ev    *fakeEvents
}

const baseSheet = `{"identity":{"name":"Ana"},"inventory":{"body":[],"backpack":[]},"assimilations":[]}`

func newRig() *rig {
	r := &rig{store: &fakeStore{}, chars: &fakeChars{sheet: json.RawMessage(baseSheet)}, rolls: &fakeRolls{},
		loot: &fakeLoot{}, ef: &fakeEffects{}, ev: &fakeEvents{}}
	toks := fakeTokens{list: []tokens.Token{
		{ID: "tok-a", SceneID: sceneID, CharacterID: charA, Name: "Ana", X: 3, Z: 3},
		{ID: "tok-npc", SceneID: sceneID, CharacterID: charNPC, Name: "Saqueador", X: 4, Z: 3},
		{ID: "tok-far", SceneID: sceneID, CharacterID: charNPC, Name: "Distante", X: 11, Z: 11},
	}}
	r.svc = NewService(r.store, fakeScenes{active: true}, toks, r.chars, r.rolls, r.loot, r.ef, r.ev)
	r.svc.pick = func(int) int { return 0 }
	return r
}

var anaSender = realtime.Sender{UserID: ownerA, Role: campaigns.RolePlayer}

func perform(r *rig, sender realtime.Sender, in PerformInput) error {
	if in.CharacterID == "" {
		in.CharacterID = charA
	}
	return r.svc.Perform(ctx, camp, sender, in)
}

func finish(r *rig, sym assimilacao.Symbols) outcome {
	r.rolls.syms = sym
	r.svc.Finalized(ctx, camp, "roll-1")
	e, _ := r.ev.last(EventResolved)
	return e.payload.(map[string]any)["outcome"].(outcome)
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
