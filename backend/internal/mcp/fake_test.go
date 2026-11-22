package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

type sent struct {
	Campaign string
	To       realtime.Filter
	Type     string
	Payload  any
}

type recNotifier struct {
	mu   sync.Mutex
	sent []sent
}

func (n *recNotifier) Broadcast(c string, to realtime.Filter, typ string, p any) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.sent = append(n.sent, sent{c, to, typ, p})
}

func (n *recNotifier) last() sent { return n.sent[len(n.sent)-1] }

// fakeScenes keeps one scene per id and records the calls it receives.
type fakeScenes struct {
	scenes    map[string]*scenes.Scene
	painted   []maps.Patch
	objects   []scenes.ObjectPatch
	origins   []string
	hit       []string
	paintErr  error
	activated []string
}

func newFakeScenes(ids ...string) *fakeScenes {
	f := &fakeScenes{scenes: map[string]*scenes.Scene{}}
	for _, id := range ids {
		f.scenes[id] = &scenes.Scene{ID: id}
	}
	return f
}

func (f *fakeScenes) List(context.Context, string) ([]scenes.Summary, error) {
	ids := make([]string, 0, len(f.scenes))
	for id := range f.scenes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]scenes.Summary, 0, len(ids))
	for _, id := range ids {
		sc := f.scenes[id]
		out = append(out, scenes.Summary{ID: id, Name: sc.Name, Width: sc.Width, Depth: sc.Depth, IsActive: sc.IsActive})
	}
	return out, nil
}

func (f *fakeScenes) Get(_ context.Context, _, id, _ string) (scenes.Scene, error) {
	sc, ok := f.scenes[id]
	if !ok {
		return scenes.Scene{}, scenes.ErrNotFound
	}
	return *sc, nil
}

func (f *fakeScenes) Paint(_ context.Context, _, sceneID, origin string, p maps.Patch) error {
	if f.paintErr != nil {
		return f.paintErr
	}
	f.hit = append(f.hit, sceneID)
	f.painted = append(f.painted, p)
	f.origins = append(f.origins, origin)
	return nil
}

func (f *fakeScenes) EditObjects(_ context.Context, _, sceneID, origin string, p scenes.ObjectPatch) error {
	f.hit = append(f.hit, sceneID)
	f.objects = append(f.objects, p)
	f.origins = append(f.origins, origin)
	return nil
}

func (f *fakeScenes) Create(_ context.Context, _, name string, w, d int) (scenes.Scene, error) {
	g, err := maps.NewGrid(w, d)
	if err != nil {
		return scenes.Scene{}, err
	}
	sc := scenes.Scene{ID: "new1", Name: name, Width: w, Depth: d, Grid: g.Serialize()}
	f.scenes[sc.ID] = &sc
	return sc, nil
}

func (f *fakeScenes) Activate(_ context.Context, _, id string) error {
	if _, ok := f.scenes[id]; !ok {
		return scenes.ErrNotFound
	}
	for k, sc := range f.scenes {
		sc.IsActive = k == id
	}
	f.activated = append(f.activated, id)
	return nil
}

type placement struct {
	Scene, Char string
	X, Z        int
}

type fakeTokens struct {
	moves    []TokenPos
	list     []tokens.Token
	placed   []placement
	placeErr map[string]error
	moveErr  error
	removed  []string
}

// List returns tokens of the scene; tokens without a SceneID belong to every scene.
func (f *fakeTokens) List(_ context.Context, _ campaigns.Member, sceneID string) ([]tokens.Token, error) {
	var out []tokens.Token
	for _, tk := range f.list {
		if tk.SceneID == "" || tk.SceneID == sceneID {
			out = append(out, tk)
		}
	}
	return out, nil
}

func (f *fakeTokens) Move(_ context.Context, _ realtime.Sender, _, id string, x, z int, force bool) error {
	if f.moveErr != nil {
		return f.moveErr
	}
	if force {
		f.moves = append(f.moves, TokenPos{id, x, z})
	}
	return nil
}

func (f *fakeTokens) Place(_ context.Context, _ realtime.Sender, _, sceneID, charID string, x, z int) error {
	if err := f.placeErr[charID]; err != nil {
		return err
	}
	f.placed = append(f.placed, placement{sceneID, charID, x, z})
	id := fmt.Sprintf("tk%d", len(f.placed))
	f.list = append(f.list, tokens.Token{ID: id, SceneID: sceneID, CharacterID: charID, Name: charID, X: x, Z: z})
	return nil
}

func (f *fakeTokens) Remove(_ context.Context, _ realtime.Sender, _, id string) error {
	for i, tk := range f.list {
		if tk.ID == id {
			f.list = append(f.list[:i], f.list[i+1:]...)
			f.removed = append(f.removed, id)
			return nil
		}
	}
	return realtime.UserError("Token não encontrado")
}

type fakeAmbience struct {
	cur        ambience.Ambience
	set        []ambience.Ambience
	conditions []ConditionState
	senders    []realtime.Sender
}

func (f *fakeAmbience) Get(context.Context, string, string) (ambience.Ambience, error) {
	return f.cur, nil
}

func (f *fakeAmbience) Set(_ context.Context, _ string, a ambience.Ambience) error {
	f.set = append(f.set, a)
	return nil
}

func (f *fakeAmbience) SetCondition(_ context.Context, snd realtime.Sender, _, ch, cond string, on bool) error {
	f.conditions = append(f.conditions, ConditionState{ch, cond, on})
	f.senders = append(f.senders, snd)
	return nil
}

type fakeCharacters struct {
	list  []characters.View
	saved map[string]json.RawMessage
}

func (f *fakeCharacters) Create(_ context.Context, _ campaigns.Member, in characters.CreateInput) (characters.View, error) {
	sheet := assimilacao.NewNPCSheet()
	sheet.Identity.Name = in.Name
	raw, _ := json.Marshal(sheet)
	v := characters.View{Character: characters.Character{
		ID: fmt.Sprintf("npc%d", len(f.list)+1), Kind: in.Kind, Name: in.Name, Sheet: raw}}
	f.list = append(f.list, v)
	return v, nil
}

func (f *fakeCharacters) Save(_ context.Context, _ campaigns.Member, id string, sheet json.RawMessage) (characters.View, error) {
	if f.saved == nil {
		f.saved = map[string]json.RawMessage{}
	}
	f.saved[id] = sheet
	return characters.View{}, nil
}

func (f *fakeCharacters) List(context.Context, campaigns.Member) ([]characters.View, error) {
	return f.list, nil
}

type rig struct {
	j   *Journal
	n   *recNotifier
	sc  *fakeScenes
	tk  *fakeTokens
	amb *fakeAmbience
	chr *fakeCharacters
	t   Table
	s   Session
}

func newRig(sceneIDs ...string) *rig {
	r := &rig{n: &recNotifier{}, sc: newFakeScenes(sceneIDs...), tk: &fakeTokens{}, amb: &fakeAmbience{}, chr: &fakeCharacters{}}
	r.j = NewJournal(r.n)
	r.t = Table{Scenes: r.sc, Tokens: r.tk, Ambience: r.amb, Characters: r.chr}
	r.s = Session{CampaignID: "c1", Sender: realtime.Sender{UserID: "gm", Role: campaigns.RoleGM, Name: "Mestre"}}
	return r
}
