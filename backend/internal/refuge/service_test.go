package refuge

import (
	"context"
	"encoding/json"
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
func (f *fakeRepo) Update(_ context.Context, _ string, d Details) (Refuge, error) {
	f.r.Details = d
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

func (f *fakeRepo) AddItem(_ context.Context, _ string, item json.RawMessage) (Refuge, error) {
	f.r.Items = append(f.r.Items, item)
	return f.r, nil
}
func (f *fakeRepo) AddResource(_ context.Context, _, _ string, amount int) (Refuge, error) {
	f.r.Water += amount
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
	repo := &fakeRepo{r: Refuge{Items: []json.RawMessage{}}, char: CharState{OwnerID: &owner, Kind: "pc", Sheet: json.RawMessage(sheet)}}
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

func TestUpdateGMOnlyAndValidation(t *testing.T) {
	s, _, _, ev := setup(sheet)
	d := Details{Name: "Refúgio", Defense: 3, Water: 5}
	if _, err := s.Update(ctx, ana, d); err == nil {
		t.Error("player updated refuge")
	}
	for _, bad := range []Details{{Name: "R", Defense: 11}, {Name: "R", Water: -1}, {Name: ""}, {Name: "R", Morale: -1}} {
		if _, err := s.Update(ctx, gm, bad); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
	if r, err := s.Update(ctx, gm, d); err != nil || r.Water != 5 || len(ev.types) != 1 {
		t.Errorf("update = %+v %v %v", r, err, ev.types)
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
	if _, err := svc.AddResource(ctx, "c", "water", 2); err != nil || repo.r.Water != 2 || len(ev.types) != 2 {
		t.Fatalf("add resource: %v %d %v", err, repo.r.Water, ev.types)
	}
}
