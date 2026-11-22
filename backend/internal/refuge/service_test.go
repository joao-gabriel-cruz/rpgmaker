package refuge

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

var ctx = context.Background()

const charID = "11111111-1111-1111-1111-111111111111"

type fakeRepo struct {
	r     Refuge
	char  CharState
	moves int
}

func (f *fakeRepo) Get(context.Context, string) (Refuge, error) { return f.r, nil }
func (f *fakeRepo) Mutate(_ context.Context, _ string, fn MutateFunc) (Refuge, error) {
	next := f.r
	next.Resources = map[string]int{}
	for k, v := range f.r.Resources {
		next.Resources[k] = v
	}
	if err := fn(&next); err != nil {
		return Refuge{}, err
	}
	f.r = next
	return f.r, nil
}
func (f *fakeRepo) Move(_ context.Context, _, _ string, fn MoveFunc) (Refuge, error) {
	sheet, items, err := fn(f.char, f.r.Items)
	if err != nil {
		return Refuge{}, err
	}
	f.char.Sheet, f.r.Items = sheet, items
	f.moves++
	return f.r, nil
}

type fakeChars struct{ announced int }

func (f *fakeChars) Announce(context.Context, string, string) error { f.announced++; return nil }

type fakeEvents struct{ types []string }

func (f *fakeEvents) Broadcast(_ string, _ realtime.Filter, typ string, _ any) {
	f.types = append(f.types, typ)
}

func setup(sheet string) (*Service, *fakeRepo, *fakeChars, *fakeEvents) {
	owner := "u-ana"
	repo := &fakeRepo{r: Refuge{Editable: Editable{Text: Text{Name: "Refúgio"}, Sheet: DefaultSheet()}, Items: []json.RawMessage{}}, char: CharState{OwnerID: &owner, Kind: "pc", Sheet: json.RawMessage(sheet)}}
	ch, ev := &fakeChars{}, &fakeEvents{}
	return NewService(repo, ch, ev), repo, ch, ev
}

var (
	ana   = campaigns.Member{CampaignID: "c", UserID: "u-ana", Role: campaigns.RolePlayer}
	bob   = campaigns.Member{CampaignID: "c", UserID: "u-bob", Role: campaigns.RolePlayer}
	gm    = campaigns.Member{CampaignID: "c", UserID: "u-gm", Role: campaigns.RoleGM}
	sheet = `{"identity":{"name":"Rita"},"inventory":{"body":[],"backpack":[{"name":"Faca","quality":1,"extra":7},{"name":"Corda"}]}}`
)

func TestDepositKeepsItemAndNotifies(t *testing.T) {
	s, repo, ch, ev := setup(sheet)
	r, err := s.Deposit(ctx, ana, charID, "backpack", 0)
	if err != nil || len(r.Items) != 1 || !strings.Contains(string(r.Items[0]), `"extra":7`) {
		t.Fatalf("deposit = %+v %v", r, err)
	}
	if strings.Contains(string(repo.char.Sheet), "Faca") || !strings.Contains(string(repo.char.Sheet), "Corda") || !strings.Contains(string(repo.char.Sheet), "Rita") {
		t.Errorf("sheet = %s", repo.char.Sheet)
	}
	if ch.announced != 1 || len(ev.types) != 1 || ev.types[0] != EventUpdated {
		t.Errorf("notify = %d %v", ch.announced, ev.types)
	}
}

func TestWithdrawRespectsLimits(t *testing.T) {
	full := `{"inventory":{"body":[{"name":"a"},{"name":"b"},{"name":"c"}],"backpack":[{},{},{},{},{},{}]}}`
	s, repo, ch, _ := setup(full)
	repo.r.Items = []json.RawMessage{json.RawMessage(`{"name":"Água"}`)}
	for slot, want := range map[string]string{"body": "corpo", "backpack": "mochila"} {
		if _, err := s.Withdraw(ctx, ana, charID, slot, 0); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s err = %v", slot, err)
		}
	}
	if len(repo.r.Items) != 1 || repo.moves != 0 || ch.announced != 0 {
		t.Errorf("state changed on failure: %+v moves=%d", repo.r.Items, repo.moves)
	}
	repo.char.Sheet = json.RawMessage(sheet)
	if _, err := s.Withdraw(ctx, ana, charID, "body", 0); err != nil || len(repo.r.Items) != 0 {
		t.Errorf("withdraw = %v", err)
	}
}

func TestMovePermissions(t *testing.T) {
	s, repo, _, _ := setup(sheet)
	if _, err := s.Deposit(ctx, bob, charID, "backpack", 0); err == nil {
		t.Error("other player moved item")
	}
	if _, err := s.Deposit(ctx, gm, charID, "backpack", 0); err != nil {
		t.Errorf("gm = %v", err)
	}
	for _, bad := range []struct {
		id, from string
		i        int
	}{{"x", "body", 0}, {charID, "pocket", 0}, {charID, "backpack", 9}} {
		if _, err := s.Deposit(ctx, ana, bad.id, bad.from, bad.i); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
	if repo.moves != 1 {
		t.Errorf("moves = %d", repo.moves)
	}
}

func editable(mod func(*Editable)) Editable {
	e := Editable{Text: Text{Name: "Refúgio"}, Sheet: DefaultSheet()}
	if mod != nil {
		mod(&e)
	}
	return e
}

func TestUpdateGMOnlyAndValidation(t *testing.T) {
	s, _, _, ev := setup(sheet)
	good := editable(func(e *Editable) { e.Defense = 3; e.Resources["water"] = 5; e.Region.Abundance["wood"] = "abundant" })
	if _, err := s.Update(ctx, ana, good); err == nil {
		t.Error("player updated refuge")
	}
	bads := []func(*Editable){
		func(e *Editable) { e.Defense = 7 }, func(e *Editable) { e.Resources["water"] = -1 },
		func(e *Editable) { e.Name = "" }, func(e *Editable) { e.Morale = -1 },
		func(e *Editable) { e.Population.Level = 11 }, func(e *Editable) { e.Population.Max = 0 },
		func(e *Editable) { e.Resources["gold"] = 1 }, func(e *Editable) { e.Region.Abundance["wood"] = "lots" },
		func(e *Editable) { e.Region.Danger = 7 }, func(e *Editable) { e.Week = 0 }, func(e *Editable) { e.Mobility = -1 },
	}
	for i, bad := range bads {
		if _, err := s.Update(ctx, gm, editable(bad)); !errors.Is(err, ErrInvalid) {
			t.Errorf("bad #%d err = %v", i, err)
		}
	}
	if r, err := s.Update(ctx, gm, good); err != nil || r.Resources["water"] != 5 || r.Derived.DefenseLabel != "Muro de madeira" || len(ev.types) != 1 {
		t.Errorf("update = %+v %v %v", r, err, ev.types)
	}
}

func TestUpdateFillsMissingKeysAndKeepsLists(t *testing.T) {
	s, repo, _, _ := setup(sheet)
	repo.r.Buildings = []json.RawMessage{json.RawMessage(`{"id":"b"}`)}
	e := Editable{Text: Text{Name: "R"}, Sheet: Sheet{Population: Population{Max: 10}, Week: 1}}
	r, err := s.Update(ctx, gm, e)
	if err != nil || len(r.Resources) != 12 || len(r.Region.Abundance) != 6 || len(r.Buildings) != 1 {
		t.Fatalf("update = %+v %v", r, err)
	}
}

func TestDerivedValues(t *testing.T) {
	sh := DefaultSheet()
	sh.Population.Level = 7
	sh.Morale, sh.Belligerence = 6, 0
	sh.Resources["water"], sh.Resources["ammo"] = 8, 5
	d := sh.Derive()
	if d.Inhabitants != 70 || d.Workers != 6 || d.ReservesTotal != 13 || d.ReservesOverflow != 3 || d.ReservesFree != 0 ||
		d.MoraleLabel != "Exaltada" || d.BelligerenceLabel != "Pacifista" || d.DefenseLabel != "Nenhuma" {
		t.Errorf("derived = %+v", d)
	}
	for lvl, want := range map[int]int{0: 0, 5: 5, 6: 5, 11: 10, 12: 10, 18: 15} {
		if got := Workers(lvl); got != want {
			t.Errorf("Workers(%d) = %d, want %d", lvl, got, want)
		}
	}
	if len(DefenseLabels) != 7 || len(MoraleLabels) != 7 || len(BelligerenceLabels) != 7 || len(ResourceKeys) != 12 {
		t.Error("label/resource tables must have 7 levels and 12 resources")
	}
}

func TestCapReservesTakesFromLargest(t *testing.T) {
	sh := DefaultSheet()
	sh.Resources["water"], sh.Resources["food"], sh.Resources["fuel"] = 9, 6, 1
	lost := sh.CapReserves()
	if sh.ReservesTotal() != 10 || sh.Resources["water"] != 4 || sh.Resources["food"] != 5 || lost["water"] != 5 || lost["food"] != 1 {
		t.Errorf("cap = %+v lost=%v", sh.Resources, lost)
	}
}

func TestAddItemAndResourceBroadcast(t *testing.T) {
	svc, repo, _, ev := setup(sheet)
	if _, err := svc.AddItem(ctx, "c", json.RawMessage(`{"name":"Pé de cabra"}`)); err != nil || len(repo.r.Items) != 1 {
		t.Fatalf("add item: %v %d", err, len(repo.r.Items))
	}
	if _, err := svc.AddResource(ctx, "c", "water", 0); err == nil {
		t.Fatal("zero amount must be rejected")
	}
	if _, err := svc.AddResource(ctx, "c", "gold", 1); err == nil {
		t.Fatal("unknown resource must be rejected")
	}
	if _, err := svc.AddResource(ctx, "c", "buildingMaterial", 2); err != nil || repo.r.Resources["buildingMaterial"] != 2 || len(ev.types) != 2 {
		t.Fatalf("add resource: %v %v %v", err, repo.r.Resources, ev.types)
	}
}

func TestAddResourceDiscardsExcess(t *testing.T) {
	svc, repo, _, _ := setup(sheet)
	repo.r.Resources["food"] = 7
	r, discarded, err := svc.AddResourceReport(ctx, "c", "ammo", 6)
	if err != nil || discarded != 3 || r.Resources["ammo"] != 3 || r.Derived.ReservesTotal != 10 {
		t.Fatalf("report = %+v %d %v", r.Resources, discarded, err)
	}
	if _, discarded, _ = svc.AddResourceReport(ctx, "c", "ammo", 2); discarded != 2 || repo.r.Resources["ammo"] != 3 {
		t.Errorf("full reserves must discard everything, got %d", discarded)
	}
}

func TestMutateRollsBackOnErrorAndValidates(t *testing.T) {
	svc, repo, _, ev := setup(sheet)
	if _, err := svc.Mutate(ctx, "c", func(r *Refuge) error { r.Week = 0; return nil }); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid mutate = %v", err)
	}
	if repo.r.Week != 1 || len(ev.types) != 0 {
		t.Errorf("state leaked: week=%d events=%v", repo.r.Week, ev.types)
	}
	r, err := svc.Mutate(ctx, "c", func(r *Refuge) error { r.Week++; return nil })
	if err != nil || r.Week != 2 || len(ev.types) != 1 {
		t.Errorf("mutate = %d %v %v", r.Week, err, ev.types)
	}
}
