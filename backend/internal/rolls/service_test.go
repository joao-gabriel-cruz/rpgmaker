package rolls

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems/assimilacao"
)

var (
	ctx    = context.Background()
	ana    = realtime.Sender{UserID: anaID, Name: anaName, Role: campaigns.RolePlayer}
	bea    = realtime.Sender{UserID: beaID, Name: "Bea", Role: campaigns.RolePlayer}
	master = realtime.Sender{UserID: gmUser, Name: "Mestre", Role: campaigns.RoleGM}
)

func newSvc(damage int, faces ...int) (*Service, *fakeRepo, *fakeEvents) {
	repo, ev := &fakeRepo{}, &fakeEvents{}
	s := NewService(repo, &fakeChars{damage: damage}, ev)
	s.rng, s.seed = &seqRNG{vals: faces}, func() int64 { return 42 }
	return s, repo, ev
}

func req(secret bool) RequestInput {
	return RequestInput{CharacterID: charID, Instinct: "reaction", Aptitude: "athletics", Secret: secret}
}

func isUserError(err error) bool {
	return fmt.Sprintf("%T", err) == "realtime.userError"
}

func TestRequestEmitsPoolForEveryone(t *testing.T) {
	s, repo, ev := newSvc(0, 0, 5, 4)
	if err := s.Request(ctx, camp, ana, req(false)); err != nil {
		t.Fatal(err)
	}
	if len(ev.got) != 1 || ev.got[0].typ != EventRolled || !ev.got[0].toPlayer || !ev.got[0].toGM {
		t.Fatalf("events = %+v", ev.got)
	}
	e := ev.got[0].payload.(Entry)
	if len(e.Dice) != 3 || e.Dice[0].Kind != "d6" || e.Dice[2].Kind != "d10" || e.Seed != 42 || e.KeptCount != 1 {
		t.Fatalf("entry = %+v", e)
	}
	if e.Dice[1].Face != 6 || e.Dice[2].Face != 5 || len(e.Suggested) != 1 || repo.recs[0].Status != StatusPending {
		t.Fatalf("faces/suggested/status wrong: %+v", e)
	}
}

func TestRequestPermissions(t *testing.T) {
	s, _, _ := newSvc(0, 1)
	if err := s.Request(ctx, camp, bea, req(false)); !isUserError(err) {
		t.Errorf("other player = %v", err)
	}
	if err := s.Request(ctx, camp, master, req(false)); err != nil {
		t.Errorf("gm for pc = %v", err)
	}
	npc := req(false)
	npc.CharacterID = npcID
	if err := s.Request(ctx, camp, master, npc); err != nil {
		t.Errorf("gm for npc = %v", err)
	}
	if err := s.Request(ctx, camp, ana, npc); !errors.Is(err, errNotFoundChar()) {
		t.Errorf("player for npc = %v", err)
	}
	bad := req(false)
	bad.Instinct = "nope"
	if err := s.Request(ctx, camp, ana, bad); !errors.As(err, new(assimilacao.RuleError)) {
		t.Errorf("bad instinct = %v", err)
	}
}

func TestSecretRollsOnlyGM(t *testing.T) {
	s, repo, ev := newSvc(0, 1)
	if err := s.Request(ctx, camp, ana, req(true)); !isUserError(err) {
		t.Fatalf("player secret = %v", err)
	}
	if err := s.Request(ctx, camp, master, req(true)); err != nil {
		t.Fatal(err)
	}
	if got := ev.got[0]; !got.toGM || got.toPlayer || repo.recs[0].Visibility != VisGM {
		t.Fatalf("secret event = %+v", got)
	}
	if err := s.Keep(ctx, camp, master, repo.recs[0].ID, []int{0}); err != nil {
		t.Fatal(err)
	}
	if got := ev.got[1]; got.typ != EventFinalized || got.toPlayer || !got.toGM {
		t.Fatalf("secret final = %+v", got)
	}
}

func TestKeepRules(t *testing.T) {
	s, repo, ev := newSvc(0, 0, 5, 4)
	_ = s.Request(ctx, camp, ana, req(false))
	id := repo.recs[0].ID
	if err := s.Keep(ctx, camp, bea, id, []int{1}); !isUserError(err) {
		t.Errorf("other user keep = %v", err)
	}
	if err := s.Keep(ctx, camp, ana, id, []int{0, 1}); !errors.As(err, new(assimilacao.RuleError)) {
		t.Errorf("wrong count = %v", err)
	}
	if err := s.Keep(ctx, camp, ana, id, []int{1}); err != nil {
		t.Fatal(err)
	}
	out := ev.got[1].payload.(map[string]any)["result"].(finalResult)
	if ev.got[1].typ != EventFinalized || out.Penalty != 0 || len(out.Kept) != 1 || out.Kept[0] != 1 {
		t.Errorf("final = %+v", ev.got[1])
	}
	if err := s.Keep(ctx, camp, ana, id, []int{1}); !errors.Is(err, ErrAlreadyFinal) {
		t.Errorf("second keep = %v", err)
	}
}

func TestKeepAppliesHealthPenalty(t *testing.T) {
	// potency 1 + resolution 1 -> 3 boxes per level; damage 6 = Laceração (-1 Success)
	s, repo, ev := newSvc(6, 0, 5, 4)
	_ = s.Request(ctx, camp, ana, req(false))
	if err := s.Keep(ctx, camp, ana, repo.recs[0].ID, []int{1}); err != nil {
		t.Fatal(err)
	}
	res := ev.got[1].payload.(map[string]any)["result"].(finalResult)
	if res.Penalty != 1 || res.HealthLevel != "Laceração" || res.Symbols.Success != assimilacao.FaceSymbols("d6", 6).Success-1 {
		t.Errorf("result = %+v", res)
	}
	var stored assimilacao.Result
	_ = json.Unmarshal(repo.recs[0].Result, &stored)
	if stored.Penalty != 1 {
		t.Errorf("stored = %+v", stored)
	}
}

func TestGenericRoll(t *testing.T) {
	s, repo, ev := newSvc(0, 1, 2) // faces 2 and 3
	if err := s.Generic(ctx, camp, bea, "2d6+3"); err != nil {
		t.Fatal(err)
	}
	e := ev.got[0].payload.(Entry)
	if e.Kind != KindGeneric || *e.Total != 8 || len(e.Dice) != 2 || e.Status != StatusFinal || !ev.got[0].toPlayer {
		t.Fatalf("entry = %+v", e)
	}
	if repo.recs[0].CharacterID != nil {
		t.Error("generic roll must have no character")
	}
	if err := s.Generic(ctx, camp, bea, "99d6"); !errors.As(err, new(RuleError)) {
		t.Errorf("invalid = %v", err)
	}
	if err := s.Keep(ctx, camp, bea, repo.recs[0].ID, nil); !errors.Is(err, ErrAlreadyFinal) {
		t.Errorf("keep generic = %v", err)
	}
}

func TestRecentHidesGMRollsFromPlayers(t *testing.T) {
	s, _, _ := newSvc(0, 1)
	_ = s.Request(ctx, camp, master, req(true))
	_ = s.Generic(ctx, camp, ana, "1d20")
	pl, _ := s.Recent(ctx, campaigns.Member{CampaignID: camp, UserID: anaID, Role: campaigns.RolePlayer}, 50)
	gm, _ := s.Recent(ctx, campaigns.Member{CampaignID: camp, UserID: gmUser, Role: campaigns.RoleGM}, 50)
	if len(pl) != 1 || len(gm) != 2 {
		t.Errorf("player=%d gm=%d", len(pl), len(gm))
	}
}
