package actions

import (
	"errors"
	"testing"
)

func TestDefaultCatalogCoversTheSpecTable(t *testing.T) {
	r := newRig()
	list, err := r.svc.Catalog(ctx, camp, false)
	if err != nil || len(list) != 11 {
		t.Fatalf("catalog = %d %v", len(list), err)
	}
	melee := list[0]
	if melee.ID != "melee_attack" || melee.Animation != "melee_strike" || melee.Outcome != OutcomeDamage ||
		len(melee.Rolls) != 2 || melee.Rolls[0].Instinct != "potency" || melee.Rolls[0].Aptitude != "weapons" {
		t.Fatalf("melee = %+v", melee)
	}
	for _, a := range list {
		if _, err := Clean(a); err != nil {
			t.Fatalf("default %s invalid: %v", a.ID, err)
		}
	}
}

func TestCampaignOverrideDisableAndCustom(t *testing.T) {
	r := newRig()
	d, _ := defaultByID("dance")
	d.Name, d.Enabled = "Dançar na roda", false
	if _, err := r.svc.Update(ctx, camp, "dance", d); err != nil {
		t.Fatal(err)
	}
	custom, err := r.svc.Create(ctx, camp, Action{Name: "Tocar tambor", Icon: "drum", Category: CatSocial, Target: TargetNone,
		Animation: "dance", Outcome: OutcomeMorale, Enabled: true, Rolls: []RollOption{{Instinct: "influence", Label: "Influência"}}})
	if err != nil || custom.Builtin || custom.ID == "" {
		t.Fatalf("create = %+v %v", custom, err)
	}
	checkDisabledAndCustom(t, r)
	if err := r.svc.Delete(ctx, camp, "dance"); err != nil {
		t.Fatal(err)
	}
	if a, err := r.svc.Find(ctx, camp, "dance"); err != nil || a.Name != "Dançar" {
		t.Fatalf("restored = %+v %v", a, err)
	}
	if err := r.svc.Delete(ctx, camp, custom.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Update(ctx, camp, "custom-nope", custom); !errors.Is(err, ErrNotFound) {
		t.Fatalf("update unknown = %v", err)
	}
}

func checkDisabledAndCustom(t *testing.T, r *rig) {
	t.Helper()
	players, _ := r.svc.Catalog(ctx, camp, false)
	gm, _ := r.svc.Catalog(ctx, camp, true)
	if len(players) != 11 || len(gm) != 12 {
		t.Fatalf("players=%d gm=%d (dance disabled, custom added)", len(players), len(gm))
	}
	if _, err := r.svc.Find(ctx, camp, "dance"); err == nil {
		t.Fatal("disabled action must not be found")
	}
}

func TestCleanRejectsBadActions(t *testing.T) {
	ok := Action{Name: "X", Icon: "eye", Category: CatCombat, Target: TargetNone, Animation: "search", Outcome: OutcomeNone,
		Rolls: []RollOption{{Instinct: "potency", Label: "L"}}}
	if _, err := Clean(ok); err != nil {
		t.Fatal(err)
	}
	for name, mut := range map[string]func(*Action){
		"name":      func(a *Action) { a.Name = " " },
		"icon":      func(a *Action) { a.Icon = "Bad Icon" },
		"category":  func(a *Action) { a.Category = "x" },
		"animation": func(a *Action) { a.Animation = "backflip" },
		"range":     func(a *Action) { a.Range = 99 },
		"no rolls":  func(a *Action) { a.Rolls = nil },
		"instinct":  func(a *Action) { a.Rolls = []RollOption{{Instinct: "luck", Label: "L"}} },
	} {
		a := ok
		mut(&a)
		if _, err := Clean(a); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}
