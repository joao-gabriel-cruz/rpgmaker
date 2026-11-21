package effects

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

const (
	camp  = "c1"
	scene = "00000000-0000-0000-0000-0000000000a1"
	charA = "00000000-0000-0000-0000-0000000000b1"
)

type fakeStore struct {
	over map[string][]Rule
	sugs []*Suggestion
	st   map[string]string
}

func newFakeStore() *fakeStore { return &fakeStore{over: map[string][]Rule{}, st: map[string]string{}} }

func (f *fakeStore) Overrides(context.Context, string) (map[string][]Rule, error) { return f.over, nil }
func (f *fakeStore) SaveRules(_ context.Context, _, k, key string, r []Rule) error {
	f.over[id(k, key)] = r
	return nil
}
func (f *fakeStore) DeleteRules(_ context.Context, _, k, key string) error {
	delete(f.over, id(k, key))
	return nil
}
func (f *fakeStore) Insert(_ context.Context, s Suggestion) (Suggestion, error) {
	s.ID = fmt.Sprintf("00000000-0000-0000-0000-%012d", len(f.sugs)+1)
	f.sugs = append(f.sugs, &s)
	f.st[s.ID] = StatusPending
	return s, nil
}
func (f *fakeStore) Pending(context.Context, string) ([]Suggestion, error) {
	out := []Suggestion{}
	for _, s := range f.sugs {
		if f.st[s.ID] == StatusPending {
			out = append(out, *s)
		}
	}
	return out, nil
}
func (f *fakeStore) Resolve(_ context.Context, _, sid, status string) (Suggestion, error) {
	for _, s := range f.sugs {
		if s.ID == sid && f.st[sid] == StatusPending {
			f.st[sid] = status
			return *s, nil
		}
	}
	return Suggestion{}, ErrNotFound
}
func (f *fakeStore) Reopen(_ context.Context, _, sid string) error {
	f.st[sid] = StatusPending
	return nil
}

type fakeWorld struct {
	sc   scenes.Scene
	toks []tokens.Token
}

func (w *fakeWorld) List(context.Context, string) ([]scenes.Summary, error) {
	return []scenes.Summary{{ID: scene, IsActive: true}}, nil
}
func (w *fakeWorld) Get(context.Context, string, string, string) (scenes.Scene, error) {
	return w.sc, nil
}

type fakeTokens struct{ w *fakeWorld }

func (f fakeTokens) Get(_ context.Context, _, tid string) (tokens.Token, error) {
	for _, t := range f.w.toks {
		if t.ID == tid {
			return t, nil
		}
	}
	return tokens.Token{}, tokens.ErrNotFound
}
func (f fakeTokens) List(context.Context, string, string) ([]tokens.Token, error) {
	return f.w.toks, nil
}

type fakeChars struct {
	sheet json.RawMessage
	owner *string
	saves int
}

func (f *fakeChars) Get(context.Context, campaigns.Member, string) (characters.View, error) {
	return characters.View{Character: characters.Character{ID: charA, OwnerID: f.owner, Sheet: f.sheet}}, nil
}
func (f *fakeChars) Save(_ context.Context, _ campaigns.Member, _ string, sh json.RawMessage) (characters.View, error) {
	f.sheet = sh
	f.saves++
	return characters.View{}, nil
}

type fakeRolls struct{ d12 int }

func (f *fakeRolls) State(context.Context, campaigns.Member, string) (rolls.State, error) {
	return rolls.State{DifficultyD12: f.d12}, nil
}
func (f *fakeRolls) SetDifficulty(_ context.Context, _ string, _ realtime.Sender, _ string, d int) error {
	f.d12 = d
	return nil
}

type sent struct {
	typ     string
	payload any
}

type fakeEvents struct {
	mu   sync.Mutex
	list []sent
}

func (e *fakeEvents) Broadcast(_ string, _ realtime.Filter, typ string, p any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.list = append(e.list, sent{typ, p})
}

func (e *fakeEvents) count(typ string) int {
	n := 0
	for _, s := range e.list {
		if s.typ == typ {
			n++
		}
	}
	return n
}

type rig struct {
	svc   *Service
	store *fakeStore
	world *fakeWorld
	chars *fakeChars
	rolls *fakeRolls
	ev    *fakeEvents
}

// newRig builds an 8x8 scene with lava at (2,2)-(3,2), snow at (5,5) and a campfire at (0,0).
func newRig() *rig {
	g, _ := maps.NewGrid(8, 8)
	for i := range g.Cells {
		g.Cells[i].Terrain = "grass"
	}
	for _, p := range [][2]int{{2, 2}, {3, 2}} {
		g.Cells[p[1]*8+p[0]].Terrain = "lava"
	}
	g.Cells[5*8+5].Terrain = "snow"
	w := &fakeWorld{sc: scenes.Scene{ID: scene, Grid: g.Serialize(), Objects: []maps.Object{{ID: "o1", Prefab: "campfire", X: 0, Z: 0}}}}
	w.toks = []tokens.Token{{ID: "t1", SceneID: scene, CharacterID: charA, Name: "Ana", X: 1, Z: 1}}
	r := &rig{store: newFakeStore(), world: w, chars: &fakeChars{sheet: json.RawMessage(`{"health":{"damage":1},"identity":{"name":"Ana"}}`)},
		rolls: &fakeRolls{}, ev: &fakeEvents{}}
	r.svc = NewService(r.store, w, fakeTokens{w}, r.chars, r.rolls, r.ev)
	return r
}

var gm = realtime.Sender{UserID: "gm", Role: campaigns.RoleGM}

func tokensAt(tid string, x, z int) tokens.Token {
	return tokens.Token{ID: tid, SceneID: scene, CharacterID: charA, Name: "N" + tid, X: x, Z: z}
}
