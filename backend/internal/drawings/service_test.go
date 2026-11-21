package drawings

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

const (
	camp    = "00000000-0000-0000-0000-00000000000c"
	sceneID = "00000000-0000-0000-0000-000000000001"
)

var (
	gm   = realtime.Sender{UserID: "gm", Name: "Mestre", Role: "gm"}
	ana  = realtime.Sender{UserID: "ana", Name: "Ana", Role: "player"}
	beto = realtime.Sender{UserID: "beto", Name: "Beto", Role: "player"}
)

type fakeRepo struct {
	mu sync.Mutex
	m  map[string]Drawing
}

func (f *fakeRepo) Add(_ context.Context, d Drawing) (Drawing, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d.ID = fmt.Sprintf("00000000-0000-0000-0000-%012d", len(f.m)+100)
	f.m[d.ID] = d
	return d, nil
}

func (f *fakeRepo) Get(_ context.Context, _, id string) (Drawing, error) {
	if d, ok := f.m[id]; ok {
		return d, nil
	}
	return Drawing{}, ErrNotFound
}

func (f *fakeRepo) List(context.Context, string, string) ([]Drawing, error) { return nil, nil }

func (f *fakeRepo) Remove(_ context.Context, id string) error {
	delete(f.m, id)
	return nil
}

func (f *fakeRepo) Clear(context.Context, string) error {
	f.m = map[string]Drawing{}
	return nil
}

type fakeScenes struct{ active bool }

func (f fakeScenes) Get(_ context.Context, _, id, role string) (scenes.Scene, error) {
	if id != sceneID || (role != "gm" && !f.active) {
		return scenes.Scene{}, scenes.ErrNotFound
	}
	return scenes.Scene{ID: id, IsActive: f.active}, nil
}

type fakeNotifier struct{ events []string }

func (n *fakeNotifier) Broadcast(_ string, _ realtime.Filter, typ string, _ any) {
	n.events = append(n.events, typ)
}

func newSvc(active bool) (*Service, *fakeRepo, *fakeNotifier) {
	r := &fakeRepo{m: map[string]Drawing{}}
	n := &fakeNotifier{}
	return NewService(r, fakeScenes{active}, n), r, n
}

func wantErr(t *testing.T, err error, sub string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), sub) {
		t.Fatalf("want error containing %q, got %v", sub, err)
	}
}

func stroke(n int) []Point {
	out := make([]Point, n)
	for i := range out {
		out[i] = Point{X: float64(i) * 0.1, Z: 1}
	}
	return out
}

func add(t *testing.T, s *Service, snd realtime.Sender) {
	t.Helper()
	if err := s.Add(context.Background(), snd, camp, sceneID, "#ff0000", 3, stroke(5)); err != nil {
		t.Fatal(err)
	}
}

func TestAddValidatesAndBroadcasts(t *testing.T) {
	s, r, n := newSvc(true)
	add(t, s, ana)
	if len(r.m) != 1 || n.events[0] != EventAdded {
		t.Fatalf("state: %v %v", r.m, n.events)
	}
	ctx := context.Background()
	wantErr(t, s.Add(ctx, ana, camp, sceneID, "red", 3, stroke(5)), "Cor")
	wantErr(t, s.Add(ctx, ana, camp, sceneID, "#ff0000", 0, stroke(5)), "Espessura")
	wantErr(t, s.Add(ctx, ana, camp, sceneID, "#ff0000", 3, stroke(1)), "curto")
	wantErr(t, s.Add(ctx, ana, camp, sceneID, "#ff0000", 3, stroke(MaxPoints+1)), "pontos demais")
	wantErr(t, s.Add(ctx, ana, camp, "nope", "#ff0000", 3, stroke(5)), "Cena")
	if err := s.Add(ctx, ana, camp, sceneID, "#ff0000", 3, stroke(MaxPoints)); err != nil {
		t.Fatal(err)
	}
}

func TestAddOnInactiveSceneDeniedForPlayer(t *testing.T) {
	s, _, _ := newSvc(false)
	wantErr(t, s.Add(context.Background(), ana, camp, sceneID, "#ff0000", 3, stroke(5)), "não está ativa")
}

func TestRemovePermissions(t *testing.T) {
	s, r, n := newSvc(true)
	ctx := context.Background()
	add(t, s, ana)
	var id string
	for k := range r.m {
		id = k
	}
	wantErr(t, s.Remove(ctx, beto, camp, id), "seus desenhos")
	if err := s.Remove(ctx, ana, camp, id); err != nil {
		t.Fatal(err)
	}
	add(t, s, ana)
	for k := range r.m {
		id = k
	}
	if err := s.Remove(ctx, gm, camp, id); err != nil || len(r.m) != 0 {
		t.Fatalf("gm remove: %v %v", err, r.m)
	}
	wantErr(t, s.Remove(ctx, gm, camp, id), "não encontrado")
	if n.events[len(n.events)-1] != EventRemoved {
		t.Fatalf("events: %v", n.events)
	}
}

func TestClearOnlyGM(t *testing.T) {
	s, r, n := newSvc(true)
	add(t, s, ana)
	wantErr(t, s.Clear(context.Background(), ana, camp, sceneID), "permissão")
	if len(r.m) != 1 {
		t.Fatal("player cleared")
	}
	if err := s.Clear(context.Background(), gm, camp, sceneID); err != nil || len(r.m) != 0 {
		t.Fatalf("clear: %v %v", err, r.m)
	}
	if n.events[len(n.events)-1] != EventCleared {
		t.Fatalf("events: %v", n.events)
	}
}

func TestRulerMoveIsEphemeral(t *testing.T) {
	s, r, n := newSvc(true)
	a, b := Point{X: 1, Z: 1}, Point{X: 3, Z: 4}
	if err := s.RulerMove(camp, ana, sceneID, &a, &b); err != nil {
		t.Fatal(err)
	}
	if err := s.RulerMove(camp, ana, sceneID, nil, nil); err != nil {
		t.Fatal(err)
	}
	wantErr(t, s.RulerMove(camp, ana, sceneID, nil, &b), "inválida")
	if len(r.m) != 0 || len(n.events) != 2 || n.events[0] != EventRuler {
		t.Fatalf("state: %v %v", r.m, n.events)
	}
}
