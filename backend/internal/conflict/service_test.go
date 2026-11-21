package conflict

import (
	"context"
	"errors"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
)

var ctx = context.Background()

const camp = "c1"

var (
	gm   = realtime.Sender{UserID: "gm", Role: campaigns.RoleGM}
	ana  = realtime.Sender{UserID: "ana", Role: campaigns.RolePlayer}
	bea  = realtime.Sender{UserID: "bea", Role: campaigns.RolePlayer}
	npcX = "npc-1"
)

type fakeRepo struct {
	cur      *Conflict
	assigned map[string]bool
}

func (f *fakeRepo) Active(context.Context, string) (*Conflict, error) {
	if f.cur == nil {
		return nil, nil
	}
	c := *f.cur
	return &c, nil
}

func (f *fakeRepo) Save(_ context.Context, c Conflict) (Conflict, error) {
	if c.ID == "" {
		c.ID = "k1"
	}
	f.cur = &c
	return c, nil
}

func (f *fakeRepo) Close(context.Context, string) error { f.cur = nil; return nil }

func (f *fakeRepo) Assign(_ context.Context, _, rollID, _ string, _ int) error {
	if f.assigned[rollID] {
		return ErrAlreadyAssigned
	}
	f.assigned[rollID] = true
	return nil
}

type fakeRolls struct {
	finals  map[string]rolls.Final
	threats []string
	sym     symbols
}

func (f *fakeRolls) Final(_ context.Context, _, id string) (rolls.Final, error) {
	r, ok := f.finals[id]
	if !ok {
		return r, rolls.ErrNotFinal
	}
	return r, nil
}

func (f *fakeRolls) Threat(_ context.Context, _ string, _ realtime.Sender, name string, _ rolls.ThreatPool) (symbols, error) {
	f.threats = append(f.threats, name)
	return f.sym, nil
}

type fakeChars struct{}

func (fakeChars) Get(_ context.Context, _ campaigns.Member, id string) (characters.View, error) {
	if id == npcX {
		return characters.View{Character: characters.Character{ID: id, Kind: characters.KindNPC}}, nil
	}
	return characters.View{Character: characters.Character{ID: id, Kind: characters.KindPC}}, nil
}

type event struct {
	typ     string
	payload any
	gm      bool
	player  bool
}

type fakeEvents struct{ evs []event }

func (f *fakeEvents) Broadcast(_ string, to realtime.Filter, typ string, p any) {
	f.evs = append(f.evs, event{typ, p, to(gm), to(ana)})
}

func (f *fakeEvents) count(typ string) int {
	n := 0
	for _, e := range f.evs {
		if e.typ == typ {
			n++
		}
	}
	return n
}

type rig struct {
	svc *Service
	rep *fakeRepo
	rl  *fakeRolls
	ev  *fakeEvents
}

func newRig() *rig {
	r := &rig{rep: &fakeRepo{assigned: map[string]bool{}}, rl: &fakeRolls{finals: map[string]rolls.Final{}}, ev: &fakeEvents{}}
	r.svc = NewService(r.rep, r.rl, fakeChars{}, r.ev)
	return r
}

func input() Input {
	pool := &rolls.ThreatPool{D6: 2, D12: 1}
	return Input{
		Title: "Cerco", Conditions: []string{" Escuro ", ""},
		Objectives:  []Objective{{ID: "o1", Title: "Fugir", Kind: KindMain, Scope: ScopeCollective, Target: 5}},
		Threats:     []Threat{{ID: "t1", Name: "Horda", Pool: pool, NpcCharacterID: &npcX}},
		Activations: []Activation{{ID: "a1", Symbol: "pressure", Threshold: 3, Text: "A horda avança", Source: SourcePlayers}},
	}
}

func (r *rig) setup(t *testing.T) {
	t.Helper()
	if err := r.svc.Save(ctx, camp, gm, input()); err != nil {
		t.Fatal(err)
	}
}

func TestSaveNormalizesAndFiltersByRole(t *testing.T) {
	r := newRig()
	r.setup(t)
	if got := r.rep.cur.Conditions; len(got) != 1 || got[0] != "Escuro" {
		t.Fatalf("conditions = %v", got)
	}
	last := r.ev.evs[len(r.ev.evs)-1]
	if last.typ != EventUpdated {
		t.Fatalf("last event %s", last.typ)
	}
	gmView, _ := r.svc.View(ctx, campaigns.Member{CampaignID: camp, Role: campaigns.RoleGM})
	plView, _ := r.svc.View(ctx, campaigns.Member{CampaignID: camp, Role: campaigns.RolePlayer})
	if gmView.Threats[0].Pool == nil || len(gmView.Activations) != 1 {
		t.Fatal("gm must see pool and activations")
	}
	if plView.Threats[0].Pool != nil || plView.Threats[0].Name != "Horda" || plView.Activations != nil || len(plView.Objectives) != 1 {
		t.Fatalf("player view leaks: %+v", plView)
	}
}

func TestSaveValidation(t *testing.T) {
	r := newRig()
	cases := map[string]func(*Input){
		"title":     func(i *Input) { i.Title = " " },
		"target":    func(i *Input) { i.Objectives[0].Target = 0 },
		"kind":      func(i *Input) { i.Objectives[0].Kind = "x" },
		"pool":      func(i *Input) { i.Threats[0].Pool = &rolls.ThreatPool{} },
		"npc":       func(i *Input) { id := "pc-1"; i.Threats[0].NpcCharacterID = &id },
		"source":    func(i *Input) { i.Activations[0].Source = "threat:nope" },
		"threshold": func(i *Input) { i.Activations[0].Threshold = 0 },
	}
	for name, mut := range cases {
		in := input()
		mut(&in)
		if err := r.svc.Save(ctx, camp, gm, in); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestAssignAddsSuccessesOnce(t *testing.T) {
	r := newRig()
	r.setup(t)
	r.rl.finals["r1"] = rolls.Final{UserID: "ana", Kind: rolls.KindAssimilacao, Symbols: symbols{Success: 2, Pressure: 1}}
	if err := r.svc.Assign(ctx, camp, ana, "r1", "o1"); err != nil {
		t.Fatal(err)
	}
	if p := r.rep.cur.Objectives[0].Progress; p != 2 {
		t.Fatalf("progress = %d", p)
	}
	if err := r.svc.Assign(ctx, camp, ana, "r1", "o1"); !errors.Is(err, ErrAlreadyAssigned) {
		t.Fatalf("second assign: %v", err)
	}
	if p := r.rep.cur.Objectives[0].Progress; p != 2 {
		t.Fatalf("progress after retry = %d", p)
	}
}

func TestAssignPermissionsAndKinds(t *testing.T) {
	r := newRig()
	r.setup(t)
	r.rl.finals["r1"] = rolls.Final{UserID: "ana", Kind: rolls.KindAssimilacao, Symbols: symbols{Success: 1}}
	r.rl.finals["th"] = rolls.Final{UserID: "gm", Kind: rolls.KindThreat, Symbols: symbols{Success: 1}}
	if err := r.svc.Assign(ctx, camp, bea, "r1", "o1"); err == nil {
		t.Fatal("other player must not assign")
	}
	if err := r.svc.Assign(ctx, camp, gm, "th", "o1"); err == nil {
		t.Fatal("threat roll must not be assignable")
	}
	if err := r.svc.Assign(ctx, camp, ana, "r1", "zz"); err == nil {
		t.Fatal("unknown objective")
	}
	if err := r.svc.Assign(ctx, camp, ana, "pending", "o1"); !errors.Is(err, rolls.ErrNotFinal) {
		t.Fatal("pending roll must fail")
	}
	if err := r.svc.Assign(ctx, camp, gm, "r1", "o1"); err != nil {
		t.Fatalf("gm may assign: %v", err)
	}
}

func TestObjectiveCompletionEvent(t *testing.T) {
	r := newRig()
	r.setup(t)
	r.rl.finals["r1"] = rolls.Final{UserID: "ana", Kind: rolls.KindAssimilacao, Symbols: symbols{Success: 5}}
	r.rl.finals["r2"] = rolls.Final{UserID: "ana", Kind: rolls.KindAssimilacao, Symbols: symbols{Success: 1}}
	_ = r.svc.Assign(ctx, camp, ana, "r1", "o1")
	_ = r.svc.Assign(ctx, camp, ana, "r2", "o1")
	if n := r.ev.count(EventObjectiveHit); n != 1 {
		t.Fatalf("completion events = %d", n)
	}
}

func TestActivationFiresForGMOnly(t *testing.T) {
	r := newRig()
	r.setup(t)
	for i, id := range []string{"r1", "r2"} {
		r.rl.finals[id] = rolls.Final{UserID: "ana", Kind: rolls.KindAssimilacao, Symbols: symbols{Pressure: 2 - i*0}}
		_ = r.svc.Assign(ctx, camp, ana, id, "o1")
	}
	if r.ev.count(EventActivation) != 1 {
		t.Fatalf("activation events = %d", r.ev.count(EventActivation))
	}
	for _, e := range r.ev.evs {
		if e.typ == EventActivation && (!e.gm || e.player) {
			t.Fatal("activation must reach only the GM")
		}
	}
	if c := r.rep.cur.Activations[0].Count; c != 1 {
		t.Fatalf("remaining count = %d", c)
	}
}

func TestRollThreat(t *testing.T) {
	r := newRig()
	in := input()
	in.Activations = append(in.Activations, Activation{ID: "a2", Symbol: "success", Threshold: 1, Text: "x", Source: "threat:t1"})
	_ = r.svc.Save(ctx, camp, gm, in)
	r.rl.sym = symbols{Success: 2}
	if err := r.svc.RollThreat(ctx, camp, gm, "t1"); err != nil {
		t.Fatal(err)
	}
	if len(r.rl.threats) != 1 || r.rl.threats[0] != "Horda" {
		t.Fatalf("threats rolled = %v", r.rl.threats)
	}
	if r.ev.count(EventThreatRolled) != 1 || r.ev.count(EventActivation) != 1 {
		t.Fatal("expected threat event and activation")
	}
	if err := r.svc.RollThreat(ctx, camp, gm, "nope"); err == nil {
		t.Fatal("unknown threat")
	}
}

func TestSavePreservesProgressAndClose(t *testing.T) {
	r := newRig()
	r.setup(t)
	r.rl.finals["r1"] = rolls.Final{UserID: "ana", Kind: rolls.KindAssimilacao, Symbols: symbols{Success: 2}}
	_ = r.svc.Assign(ctx, camp, ana, "r1", "o1")
	in := input()
	in.Title = "Novo"
	in.Objectives[0].Progress = 99
	_ = r.svc.Save(ctx, camp, gm, in)
	if r.rep.cur.Objectives[0].Progress != 2 {
		t.Fatal("progress must be server-owned")
	}
	_ = r.svc.Close(ctx, camp)
	if v, _ := r.svc.View(ctx, campaigns.Member{CampaignID: camp, Role: campaigns.RoleGM}); v != nil {
		t.Fatal("closed conflict must be gone")
	}
	if err := r.svc.RollThreat(ctx, camp, gm, "t1"); err == nil {
		t.Fatal("no active conflict")
	}
}

type noEvents struct{}

func (noEvents) Broadcast(string, realtime.Filter, string, any) {}

func senderOf(id string) realtime.Sender { return realtime.Sender{UserID: id, Role: campaigns.RoleGM} }
