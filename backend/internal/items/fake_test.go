package items

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/refuge"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

const (
	camp   = "c1"
	sceneA = "00000000-0000-0000-0000-0000000000a1"
	charA  = "00000000-0000-0000-0000-0000000000b1"
	owner  = "u-ana"
)

type fakeRepo struct {
	items map[string]Item
	loot  map[string]Loot
	n     int
}

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[string]Item{}, loot: map[string]Loot{}} }

func (f *fakeRepo) id() string { f.n++; return fmt.Sprintf("00000000-0000-0000-0000-%012d", f.n) }

func (f *fakeRepo) ListItems(context.Context, string) ([]Item, error) {
	out := []Item{}
	for _, it := range f.items {
		out = append(out, it)
	}
	return out, nil
}
func (f *fakeRepo) GetItem(_ context.Context, _, id string) (Item, error) {
	if it, ok := f.items[id]; ok {
		return it, nil
	}
	return Item{}, ErrNotFound
}
func (f *fakeRepo) CreateItem(_ context.Context, _ string, it Item) (Item, error) {
	it.ID = f.id()
	f.items[it.ID] = it
	return it, nil
}
func (f *fakeRepo) UpdateItem(_ context.Context, _ string, it Item) (Item, error) {
	if _, ok := f.items[it.ID]; !ok {
		return Item{}, ErrNotFound
	}
	f.items[it.ID] = it
	return it, nil
}
func (f *fakeRepo) DeleteItem(_ context.Context, _, id string) error {
	delete(f.items, id)
	return nil
}
func (f *fakeRepo) InsertLoot(_ context.Context, l Loot) (Loot, error) {
	l.ID = f.id()
	f.loot[l.ID] = l
	return l, nil
}
func (f *fakeRepo) ListLoot(_ context.Context, _, sceneID string) ([]Loot, error) {
	out := []Loot{}
	for _, l := range f.loot {
		if l.SceneID == sceneID {
			out = append(out, l)
		}
	}
	return out, nil
}
func (f *fakeRepo) GetLoot(_ context.Context, _, id string) (Loot, error) {
	if l, ok := f.loot[id]; ok {
		return l, nil
	}
	return Loot{}, ErrNotFound
}
func (f *fakeRepo) SaveLoot(_ context.Context, l Loot) error { f.loot[l.ID] = l; return nil }
func (f *fakeRepo) DeleteLoot(_ context.Context, _, id string) error {
	delete(f.loot, id)
	return nil
}

type fakeChars struct{ sheet json.RawMessage }

func (f *fakeChars) Get(_ context.Context, _ campaigns.Member, id string) (characters.View, error) {
	if id != charA {
		return characters.View{}, characters.ErrNotFound
	}
	o := owner
	return characters.View{Character: characters.Character{ID: id, Name: "Kaio", OwnerID: &o, Kind: "pc", Sheet: f.sheet}}, nil
}
func (f *fakeChars) EditSheet(_ context.Context, _, id string, fn func(json.RawMessage) (json.RawMessage, error)) (characters.View, error) {
	if id != charA {
		return characters.View{}, characters.ErrNotFound
	}
	out, err := fn(f.sheet)
	if err != nil {
		return characters.View{}, err
	}
	f.sheet = out
	return characters.View{}, nil
}

func (f *fakeChars) slot(place string) []map[string]any {
	var doc struct {
		Inventory map[string][]map[string]any `json:"inventory"`
	}
	_ = json.Unmarshal(f.sheet, &doc)
	return doc.Inventory[place]
}

type fakeRefuge struct {
	items []json.RawMessage
	water int
}

func (f *fakeRefuge) AddItem(_ context.Context, _ string, it json.RawMessage) (refuge.Refuge, error) {
	f.items = append(f.items, it)
	return refuge.Refuge{}, nil
}
func (f *fakeRefuge) AddResource(_ context.Context, _, _ string, n int) (refuge.Refuge, error) {
	f.water += n
	return refuge.Refuge{}, nil
}

type fakeEffects struct{ got []effects.Suggestion }

func (f *fakeEffects) Suggest(_ context.Context, in effects.Suggestion) (effects.Suggestion, error) {
	f.got = append(f.got, in)
	return in, nil
}

type fakeTokens struct{ x, z int }

func (f *fakeTokens) List(context.Context, campaigns.Member, string) ([]tokens.Token, error) {
	return []tokens.Token{{CharacterID: charA, SceneID: sceneA, X: f.x, Z: f.z}}, nil
}

type fakeScenes struct{}

func (fakeScenes) Get(_ context.Context, _, id, _ string) (scenes.Scene, error) {
	if id != sceneA {
		return scenes.Scene{}, scenes.ErrNotFound
	}
	return scenes.Scene{ID: id, Width: 10, Depth: 10, Objects: []maps.Object{{ID: "crate1", Prefab: "crate"}}}, nil
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

// playerSees lists the event types a player would receive.
func (f *fakeEvents) playerSees() []string {
	var out []string
	for _, e := range f.list {
		if e.to(realtime.Sender{Role: campaigns.RolePlayer, UserID: "u-other"}) {
			out = append(out, e.typ)
		}
	}
	return out
}

type rig struct {
	svc   *Service
	repo  *fakeRepo
	chars *fakeChars
	rf    *fakeRefuge
	ef    *fakeEffects
	tk    *fakeTokens
	ev    *fakeEvents
}

const emptySheet = `{"identity":{"name":"Kaio"},"inventory":{"body":[],"backpack":[]}}`

func newRig() *rig {
	r := &rig{repo: newFakeRepo(), chars: &fakeChars{sheet: json.RawMessage(emptySheet)}, rf: &fakeRefuge{}, ef: &fakeEffects{}, tk: &fakeTokens{}, ev: &fakeEvents{}}
	r.svc = NewService(r.repo, r.chars, r.rf, r.ef, r.tk, fakeScenes{}, r.ev)
	return r
}

func validItem() Item {
	return Item{Name: "Pé de cabra", Icon: "wrench", Color: "#aabbcc", Kind: KindTool, Quality: 1, Traits: []string{"metal"},
		Effects: ItemEffects{KeptBonus: &KeptBonus{ActionCategory: "exploration", Amount: 1}}}
}
