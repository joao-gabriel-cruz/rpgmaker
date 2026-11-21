package mcp

import (
	"context"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/ambience"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
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
	scenes  map[string]*scenes.Scene
	painted []maps.Patch
	objects []scenes.ObjectPatch
	origins []string
}

func newFakeScenes(ids ...string) *fakeScenes {
	f := &fakeScenes{scenes: map[string]*scenes.Scene{}}
	for _, id := range ids {
		f.scenes[id] = &scenes.Scene{ID: id}
	}
	return f
}

func (f *fakeScenes) List(context.Context, string) ([]scenes.Summary, error) { return nil, nil }

func (f *fakeScenes) Get(_ context.Context, _, id, _ string) (scenes.Scene, error) {
	sc, ok := f.scenes[id]
	if !ok {
		return scenes.Scene{}, scenes.ErrNotFound
	}
	return *sc, nil
}

func (f *fakeScenes) Paint(_ context.Context, _, _, origin string, p maps.Patch) error {
	f.painted = append(f.painted, p)
	f.origins = append(f.origins, origin)
	return nil
}

func (f *fakeScenes) EditObjects(_ context.Context, _, _, origin string, p scenes.ObjectPatch) error {
	f.objects = append(f.objects, p)
	f.origins = append(f.origins, origin)
	return nil
}

func (f *fakeScenes) Create(context.Context, string, string, int, int) (scenes.Scene, error) {
	return scenes.Scene{}, nil
}

func (f *fakeScenes) Activate(context.Context, string, string) error { return nil }

type fakeTokens struct{ moves []TokenPos }

func (f *fakeTokens) List(context.Context, campaigns.Member, string) ([]tokens.Token, error) {
	return nil, nil
}

func (f *fakeTokens) Move(_ context.Context, _ realtime.Sender, _, id string, x, z int, force bool) error {
	if force {
		f.moves = append(f.moves, TokenPos{id, x, z})
	}
	return nil
}

func (f *fakeTokens) Place(context.Context, realtime.Sender, string, string, string, int, int) error {
	return nil
}

type fakeAmbience struct {
	set        []ambience.Ambience
	conditions []ConditionState
	senders    []realtime.Sender
}

func (f *fakeAmbience) Get(context.Context, string, string) (ambience.Ambience, error) {
	return ambience.Ambience{}, nil
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

type fakeCharacters struct{}

func (fakeCharacters) List(context.Context, campaigns.Member) ([]characters.View, error) {
	return nil, nil
}

type rig struct {
	j   *Journal
	n   *recNotifier
	sc  *fakeScenes
	tk  *fakeTokens
	amb *fakeAmbience
	t   Table
	s   Session
}

func newRig(sceneIDs ...string) *rig {
	r := &rig{n: &recNotifier{}, sc: newFakeScenes(sceneIDs...), tk: &fakeTokens{}, amb: &fakeAmbience{}}
	r.j = NewJournal(r.n)
	r.t = Table{Scenes: r.sc, Tokens: r.tk, Ambience: r.amb, Characters: fakeCharacters{}}
	r.s = Session{CampaignID: "c1", Sender: realtime.Sender{UserID: "gm", Role: campaigns.RoleGM, Name: "Mestre"}}
	return r
}
