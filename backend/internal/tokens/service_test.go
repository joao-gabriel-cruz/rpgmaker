package tokens

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

const (
	camp    = "00000000-0000-0000-0000-00000000000c"
	sceneID = "00000000-0000-0000-0000-000000000001"
	charA   = "00000000-0000-0000-0000-0000000000a1"
	charB   = "00000000-0000-0000-0000-0000000000b1"
)

var (
	gm   = realtime.Sender{UserID: "gm", Name: "Mestre", Role: "gm"}
	ana  = realtime.Sender{UserID: "ana", Name: "Ana", Role: "player"}
	beto = realtime.Sender{UserID: "beto", Name: "Beto", Role: "player"}
)

type fakeRepo struct {
	mu     sync.Mutex
	tokens map[string]*Token
	chars  map[string]Char
}

func (f *fakeRepo) Place(_ context.Context, sc, ch string, x, z int) (Token, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, t := range f.tokens {
		if t.SceneID == sc && t.CharacterID == ch {
			return Token{}, ErrExists
		}
	}
	id := fmt.Sprintf("00000000-0000-0000-0000-%012d", len(f.tokens)+100)
	t := &Token{ID: id, SceneID: sc, CharacterID: ch, X: x, Z: z, OwnerID: f.chars[ch].OwnerID}
	f.tokens[id] = t
	return *t, nil
}

func (f *fakeRepo) Get(_ context.Context, _, id string) (Token, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.tokens[id]; ok {
		return *t, nil
	}
	return Token{}, ErrNotFound
}

func (f *fakeRepo) List(context.Context, string, string) ([]Token, error) { return nil, nil }

func (f *fakeRepo) Move(_ context.Context, id string, x, z int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokens[id].X, f.tokens[id].Z = x, z
	return nil
}

func (f *fakeRepo) Remove(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.tokens, id)
	return nil
}

func (f *fakeRepo) Character(_ context.Context, _, id string) (Char, error) {
	if c, ok := f.chars[id]; ok {
		return c, nil
	}
	return Char{}, ErrNotFound
}

type fakeScenes struct{ sc scenes.Scene }

func (f fakeScenes) Get(_ context.Context, _, id, role string) (scenes.Scene, error) {
	if id != f.sc.ID || (role != "gm" && !f.sc.IsActive) {
		return scenes.Scene{}, scenes.ErrNotFound
	}
	return f.sc, nil
}

type fakeNotifier struct{ events []string }

func (n *fakeNotifier) Broadcast(_ string, _ realtime.Filter, typ string, _ any) {
	n.events = append(n.events, typ)
}

func str(s string) *string { return &s }

// newSvc builds a 16x16 scene with a wall column at x=5 and a 3-high step at x=9.
func newSvc(t *testing.T) (*Service, *fakeRepo, *fakeNotifier) {
	t.Helper()
	g, _ := maps.NewGrid(16, 16)
	for z := 0; z < 16; z++ {
		g.Cells[z*16+9].Height = 4
	}
	sc := scenes.Scene{ID: sceneID, Width: 16, Depth: 16, Grid: g.Serialize(), IsActive: true,
		Objects: wallColumn()}
	repo := &fakeRepo{tokens: map[string]*Token{}, chars: map[string]Char{
		charA: {ID: charA, Name: "A", OwnerID: str("ana")}, charB: {ID: charB, Name: "B", OwnerID: str("beto")}}}
	n := &fakeNotifier{}
	return NewService(repo, fakeScenes{sc}, n), repo, n
}

func wallColumn() []maps.Object {
	var out []maps.Object
	for z := 0; z < 16; z++ {
		out = append(out, maps.Object{ID: fmt.Sprintf("w%d", z), Prefab: "wall", X: 5, Z: z})
	}
	return out
}

func wantErr(t *testing.T, err error, sub string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), sub) {
		t.Fatalf("want error containing %q, got %v", sub, err)
	}
}

func place(t *testing.T, s *Service, snd realtime.Sender, ch string, x, z int) {
	t.Helper()
	if err := s.Place(context.Background(), snd, camp, sceneID, ch, x, z); err != nil {
		t.Fatal(err)
	}
}

func tokenOf(r *fakeRepo, ch string) *Token {
	for _, t := range r.tokens {
		if t.CharacterID == ch {
			return t
		}
	}
	return nil
}
