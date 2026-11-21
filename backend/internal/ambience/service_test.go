package ambience

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

type memStore map[string]Ambience

func (m memStore) Get(_ context.Context, id string) (Ambience, error) {
	if a, ok := m[id]; ok {
		return a, nil
	}
	return Default(id), nil
}
func (m memStore) Save(_ context.Context, a Ambience) error { m[a.SceneID] = a; return nil }

type fakeScenes struct{}

func (fakeScenes) Get(_ context.Context, _, id, _ string) (scenes.Scene, error) {
	if id != "s1" {
		return scenes.Scene{}, scenes.ErrNotFound
	}
	return scenes.Scene{}, nil
}

type fakeChars struct{ sheet json.RawMessage }

func (f *fakeChars) Get(_ context.Context, _ campaigns.Member, _ string) (characters.View, error) {
	owner := "u-ana"
	return characters.View{Character: characters.Character{ID: "c1", Name: "Ana", OwnerID: &owner}}, nil
}

func (f *fakeChars) EditSheet(_ context.Context, _, _ string, fn SheetEdit) (characters.View, error) {
	var err error
	f.sheet, err = fn(f.sheet)
	return characters.View{}, err
}

type sent struct {
	typ     string
	payload any
}
type fakeEvents struct{ out []sent }

func (e *fakeEvents) Broadcast(_ string, _ realtime.Filter, typ string, p any) {
	e.out = append(e.out, sent{typ, p})
}

func newSvc() (*Service, memStore, *fakeChars, *fakeEvents) {
	st, ch, ev := memStore{}, &fakeChars{sheet: json.RawMessage(`{"notes":"x"}`)}, &fakeEvents{}
	return NewService(st, fakeScenes{}, ch, ev), st, ch, ev
}

func TestSetPersistsAndBroadcasts(t *testing.T) {
	svc, st, _, ev := newSvc()
	a := Ambience{SceneID: "s1", TimeOfDay: "night", Weather: "rain"}
	if err := svc.Set(context.Background(), "c", a); err != nil {
		t.Fatal(err)
	}
	if st["s1"] != a || len(ev.out) != 1 || ev.out[0].typ != EventUpdated {
		t.Fatalf("state %v events %v", st, ev.out)
	}
	for _, bad := range []Ambience{{"s1", "noon", "rain"}, {"s1", "day", "hail"}, {"zz", "day", "clear"}} {
		if svc.Set(context.Background(), "c", bad) == nil {
			t.Fatalf("%v must be rejected", bad)
		}
	}
}

func TestConditionPermissions(t *testing.T) {
	svc, _, ch, _ := newSvc()
	ctx := context.Background()
	gm := realtime.Sender{UserID: "gm", Role: campaigns.RoleGM}
	ana := realtime.Sender{UserID: "u-ana", Role: campaigns.RolePlayer}
	if svc.SetCondition(ctx, ana, "c", "c1", "infected", true) == nil {
		t.Fatal("owner may not infect themselves")
	}
	if err := svc.SetCondition(ctx, ana, "c", "c1", "hidden", true); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetCondition(ctx, gm, "c", "c1", "scared", true); err != nil {
		t.Fatal(err)
	}
	if string(ch.sheet) != `{"conditions":["hidden","scared"],"notes":"x"}` {
		t.Fatalf("sheet %s", ch.sheet)
	}
	_ = svc.SetCondition(ctx, gm, "c", "c1", "hidden", false)
	_ = svc.SetCondition(ctx, gm, "c", "c1", "scared", false)
	if string(ch.sheet) != `{"notes":"x"}` {
		t.Fatalf("empty list must drop the key: %s", ch.sheet)
	}
}

func TestUnknownConditionRejected(t *testing.T) {
	svc, _, _, _ := newSvc()
	if svc.SetCondition(context.Background(), realtime.Sender{Role: campaigns.RoleGM}, "c", "c1", "sleepy", true) == nil {
		t.Fatal("unknown condition")
	}
}

func TestChat(t *testing.T) {
	svc, _, _, ev := newSvc()
	ctx := context.Background()
	ana := realtime.Sender{UserID: "u-ana", Role: campaigns.RolePlayer}
	bob := realtime.Sender{UserID: "u-bob", Role: campaigns.RolePlayer}
	if err := svc.Say(ctx, ana, "c", "c1", "  Olá  "); err != nil {
		t.Fatal(err)
	}
	if s := ev.out[0].payload.(said); s.Text != "Olá" || s.Name != "Ana" {
		t.Fatalf("%+v", s)
	}
	long := make([]rune, MaxSayLen+1)
	for i := range long {
		long[i] = 'a'
	}
	if svc.Say(ctx, ana, "c", "c1", string(long)) == nil || svc.Say(ctx, ana, "c", "c1", " ") == nil {
		t.Fatal("empty and long texts must fail")
	}
	if svc.Say(ctx, bob, "c", "c1", "oi") == nil {
		t.Fatal("only the owner or Mestre speaks")
	}
	if svc.React(ctx, ana, "c", "c1", "💀") != nil || svc.React(ctx, ana, "c", "c1", "x") == nil {
		t.Fatal("emoji whitelist")
	}
}
