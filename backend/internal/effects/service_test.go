package effects

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rounds"
)

var bg = context.Background()

func pending(t *testing.T, r *rig) []Suggestion {
	t.Helper()
	list, _ := r.svc.Pending(bg, camp)
	return list
}

func TestEnterLavaSuggestsOnEnterOnly(t *testing.T) {
	r := newRig()
	r.svc.OnEntered(bg, camp, "t1", 2, 2, &maps.Coord{X: 1, Z: 1})
	list := pending(t, r)
	if len(list) != 1 || list[0].Source != "Lava" || list[0].Effect != EffectDamage || list[0].Amount != 2 || list[0].CharacterName != "Ana" {
		t.Fatalf("got %+v", list)
	}
	if r.ev.count(EventSuggested) != 1 {
		t.Fatal("expected one effect.suggested event")
	}
}

func TestMoveWithinSameSourceIsSilent(t *testing.T) {
	r := newRig()
	r.svc.OnEntered(bg, camp, "t1", 3, 2, &maps.Coord{X: 2, Z: 2})
	if n := len(pending(t, r)); n != 0 {
		t.Fatalf("lava to lava must not retrigger, got %d", n)
	}
	r.svc.OnEntered(bg, camp, "t1", 2, 2, &maps.Coord{X: 4, Z: 4})
	r.svc.OnEntered(bg, camp, "t1", 2, 2, nil)
	if n := len(pending(t, r)); n != 2 {
		t.Fatalf("re-entry and placement must trigger, got %d", n)
	}
}

func TestEnterPrefabObjectCell(t *testing.T) {
	r := newRig()
	r.svc.OnEntered(bg, camp, "t1", 0, 0, &maps.Coord{X: 1, Z: 1})
	list := pending(t, r)
	if len(list) != 1 || list[0].Source != "Fogueira" || list[0].Amount != 1 {
		t.Fatalf("got %+v", list)
	}
}

func TestTurnStartHookOnNewPreparationOnly(t *testing.T) {
	r := newRig()
	r.world.toks = append(r.world.toks, tokensAt("t2", 2, 2), tokensAt("t3", 5, 5), tokensAt("t4", 7, 7))
	idle, start := rounds.Idle(), rounds.Start()
	r.svc.StartOfTurn(bg, camp, start, rounds.State{Active: true, Round: 1, Phase: rounds.PhaseInfected})
	if len(pending(t, r)) != 0 {
		t.Fatal("mid-round phase must not trigger")
	}
	r.svc.StartOfTurn(bg, camp, idle, start)
	list := pending(t, r)
	if len(list) != 2 {
		t.Fatalf("want lava + snow, got %+v", list)
	}
	r.svc.StartOfTurn(bg, camp, rounds.State{Active: true, Round: 1, Phase: rounds.PhaseEnd}, rounds.State{Active: true, Round: 2, Phase: rounds.PhasePreparation})
	if len(pending(t, r)) != 4 {
		t.Fatal("new round must trigger again")
	}
	r.svc.StartOfTurn(bg, camp, start, idle)
	if len(pending(t, r)) != 4 {
		t.Fatal("ending the conflict must not trigger")
	}
}

func TestConfirmDamageUpdatesSheet(t *testing.T) {
	r := newRig()
	owner := "u-ana"
	r.chars.owner = &owner
	r.svc.OnEntered(bg, camp, "t1", 2, 2, nil)
	sg := pending(t, r)[0]
	if err := r.svc.Confirm(bg, camp, gm, sg.ID); err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Health   struct{ Damage int }
		Identity map[string]string
	}
	_ = json.Unmarshal(r.chars.sheet, &doc)
	if doc.Health.Damage != 3 || doc.Identity["name"] != "Ana" {
		t.Fatalf("sheet = %s", r.chars.sheet)
	}
	if r.ev.count(EventResolved) != 1 || r.ev.count(EventApplied) != 1 || len(pending(t, r)) != 0 {
		t.Fatalf("events %+v", r.ev.list)
	}
	if err := r.svc.Confirm(bg, camp, gm, sg.ID); err == nil {
		t.Fatal("second confirm must fail, damage applies once")
	}
}

func TestConfirmDifficultyAccumulatesWithCap(t *testing.T) {
	r := newRig()
	r.svc.OnEntered(bg, camp, "t1", 5, 5, nil)
	for i := 0; i < 4; i++ {
		r.svc.OnTurnStart(bg, camp)
		r.world.toks[0].X, r.world.toks[0].Z = 5, 5
	}
	for _, sg := range pending(t, r) {
		if sg.Effect != EffectDifficulty {
			continue
		}
		if err := r.svc.Confirm(bg, camp, gm, sg.ID); err != nil {
			t.Fatal(err)
		}
	}
	if r.rolls.d12 != 3 {
		t.Fatalf("d12 = %d, want cap 3", r.rolls.d12)
	}
}

func TestDismissAndNoticeChangeNothing(t *testing.T) {
	r := newRig()
	r.svc.OnEntered(bg, camp, "t1", 2, 2, nil)
	if err := r.svc.Dismiss(bg, camp, pending(t, r)[0].ID); err != nil {
		t.Fatal(err)
	}
	if r.chars.saves != 0 || r.rolls.d12 != 0 || len(pending(t, r)) != 0 || r.ev.count(EventResolved) != 1 {
		t.Fatal("dismiss must only close the suggestion")
	}
	if err := r.svc.Dismiss(bg, camp, "not-a-uuid"); err == nil {
		t.Fatal("invalid id must fail")
	}
}

func TestNoticeConfirmOnlyRecords(t *testing.T) {
	r := newRig()
	r.svc.suggest(bg, camp, r.world.toks[0], 4, 4, TriggerEnter, []source{{"terrain:water", KindTerrain, "water"}})
	if err := r.svc.Confirm(bg, camp, gm, pending(t, r)[0].ID); err != nil || r.chars.saves != 0 {
		t.Fatalf("notice must not touch the sheet: %v", err)
	}
}

func TestRulesOverrideAndRestore(t *testing.T) {
	r := newRig()
	set, err := r.svc.SetRules(bg, camp, KindTerrain, "lava", []Rule{{TriggerEnter, EffectDamage, 5, " Brasa "}})
	if err != nil || set.IsDefault || set.Rules[0].Message != "Brasa" {
		t.Fatalf("%+v %v", set, err)
	}
	r.svc.OnEntered(bg, camp, "t1", 2, 2, nil)
	if sg := pending(t, r); len(sg) != 1 || sg[0].Amount != 5 {
		t.Fatalf("override ignored: %+v", sg)
	}
	set, _ = r.svc.RestoreRules(bg, camp, KindTerrain, "lava")
	if !set.IsDefault || len(set.Rules) != 2 {
		t.Fatalf("restore: %+v", set)
	}
	if _, err := r.svc.SetRules(bg, camp, KindTerrain, "lava", nil); err != nil {
		t.Fatal("empty rules disable the effects")
	}
}

func TestRulesValidation(t *testing.T) {
	r := newRig()
	long := strings.Repeat("x", MaxMessageLen+1)
	bad := [][]Rule{
		{{"sometimes", EffectDamage, 1, ""}}, {{TriggerEnter, "heal", 1, ""}}, {{TriggerEnter, EffectDamage, 0, ""}},
		{{TriggerEnter, EffectDifficulty, 4, ""}}, {{TriggerEnter, EffectNotice, 1, ""}}, {{TriggerEnter, EffectNotice, 0, long}},
	}
	for _, b := range bad {
		if _, err := r.svc.SetRules(bg, camp, KindTerrain, "lava", b); err == nil {
			t.Fatalf("accepted %+v", b)
		}
	}
	if _, err := r.svc.SetRules(bg, camp, "wizard", "lava", nil); err == nil {
		t.Fatal("unknown kind accepted")
	}
	list, _ := r.svc.Rules(bg, camp)
	if len(list) != len(maps.TerrainOrder)+len(maps.PrefabIDs()) {
		t.Fatalf("listing size %d", len(list))
	}
	for _, rs := range list {
		if rs.Label == rs.Key {
			t.Fatalf("missing label for %s:%s", rs.Kind, rs.Key)
		}
	}
}

func TestConfirmHealReducesDamageWithoutScene(t *testing.T) {
	r := newRig()
	sg, err := r.svc.Suggest(bg, Suggestion{CampaignID: camp, CharacterID: charA, CharacterName: "Ana", Source: "Bandagem",
		Trigger: TriggerItemUse, Effect: EffectHeal, Amount: 5, Message: "Usou Bandagem: curar 5"})
	if err != nil || r.ev.count(EventSuggested) != 1 {
		t.Fatalf("suggest: %v", err)
	}
	if err := r.svc.Confirm(bg, camp, gm, sg.ID); err != nil {
		t.Fatal(err)
	}
	var doc struct{ Health struct{ Damage int } }
	_ = json.Unmarshal(r.chars.sheet, &doc)
	if doc.Health.Damage != 0 {
		t.Fatalf("damage must floor at 0, sheet = %s", r.chars.sheet)
	}
}
