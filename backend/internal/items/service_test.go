package items

import (
	"context"
	"strings"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

var ctx = context.Background()

var invalidItems = map[string]func(*Item){
	"nome":        func(i *Item) { i.Name = " " },
	"Ícone":       func(i *Item) { i.Icon = "nope" },
	"Cor":         func(i *Item) { i.Color = "red" },
	"Tipo":        func(i *Item) { i.Kind = "x" },
	"0 a 10":      func(i *Item) { i.Quality = 11 },
	"mantido":     func(i *Item) { i.Effects.KeptBonus.Amount = 3 },
	"cura":        func(i *Item) { i.Kind = KindConsumable; i.Effects = ItemEffects{Heal: 7} },
	"consumíveis": func(i *Item) { i.Effects = ItemEffects{Heal: 2} },
	"animação":    func(i *Item) { i.AttackAnimation = "slash" },
	"Recurso": func(i *Item) {
		i.Kind = KindConsumable
		i.Effects = ItemEffects{RefugeResource: &RefugeResource{Resource: "gold", Amount: 1}}
	},
}

func TestCleanValidatesItem(t *testing.T) {
	for want, mutate := range invalidItems {
		it := validItem()
		mutate(&it)
		if _, err := Clean(it); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err = %v", want, err)
		}
	}
	if _, err := Clean(validItem()); err != nil {
		t.Fatal(err)
	}
	sword := Item{Name: "Faca", Icon: "slice", Color: "#000000", Kind: KindMelee, AttackAnimation: "slash"}
	if _, err := Clean(sword); err != nil {
		t.Fatal(err)
	}
}

func TestCompendiumCRUDNotifiesGM(t *testing.T) {
	r := newRig()
	it, err := r.svc.Create(ctx, camp, validItem())
	if err != nil || it.ID == "" {
		t.Fatalf("create: %v", err)
	}
	it.Name = "Barra"
	if out, err := r.svc.Update(ctx, camp, it.ID, it); err != nil || out.Name != "Barra" {
		t.Fatalf("update: %v", err)
	}
	if err := r.svc.Delete(ctx, camp, it.ID); err != nil {
		t.Fatal(err)
	}
	if len(r.ev.list) != 3 || len(r.ev.playerSees()) != 0 {
		t.Fatalf("events = %+v", r.ev.list)
	}
}

func giveTo(r *rig, id, place string, qty int) error {
	return r.svc.Give(ctx, camp, GiveInput{ItemID: id, CharacterID: charA, Place: place, Qty: qty})
}

func TestGiveRespectsLimitsAndKeepsCompendiumFields(t *testing.T) {
	r := newRig()
	it, _ := r.svc.Create(ctx, camp, validItem())
	if err := giveTo(r, it.ID, PlaceBackpack, 1); err != nil {
		t.Fatal(err)
	}
	got := r.chars.slot("backpack")
	if len(got) != 1 || got[0]["itemId"] != it.ID || got[0]["icon"] != "wrench" || got[0]["name"] != "Pé de cabra" {
		t.Fatalf("backpack = %+v", got)
	}
	if err := giveTo(r, it.ID, PlaceBody, 4); err == nil || !strings.Contains(err.Error(), "corpo") {
		t.Fatalf("body overflow err = %v", err)
	}
	if len(r.chars.slot("body")) != 0 {
		t.Fatal("all-or-nothing: nothing may be given on overflow")
	}
	if err := giveTo(r, it.ID, PlaceBackpack, 6); err == nil {
		t.Fatal("backpack must be full after 1+6")
	}
	if err := giveTo(r, it.ID, "pocket", 1); err == nil {
		t.Fatal("bad place accepted")
	}
}

func TestGiveToRefugeAndUnknownTargets(t *testing.T) {
	r := newRig()
	it, _ := r.svc.Create(ctx, camp, validItem())
	if err := r.svc.Give(ctx, camp, GiveInput{ItemID: it.ID, Target: TargetRefuge, Qty: 2}); err != nil || len(r.rf.items) != 2 {
		t.Fatalf("refuge give: %v %d", err, len(r.rf.items))
	}
	if err := r.svc.Give(ctx, camp, GiveInput{ItemID: "nope", CharacterID: charA, Place: PlaceBody}); err == nil {
		t.Fatal("unknown item accepted")
	}
	if err := r.svc.Give(ctx, camp, GiveInput{ItemID: it.ID, CharacterID: "00000000-0000-0000-0000-00000000dead", Place: PlaceBody}); err == nil {
		t.Fatal("unknown character accepted")
	}
}

func bandage(r *rig) Item {
	it, _ := r.svc.Create(ctx, camp, Item{Name: "Bandagem", Icon: "bandage", Color: "#ffffff", Kind: KindConsumable, Effects: ItemEffects{Heal: 2}})
	return it
}

func TestUseConsumableSuggestsHeal(t *testing.T) {
	r := newRig()
	b := bandage(r)
	_ = giveTo(r, b.ID, PlaceBackpack, 1)
	ana := campaigns.Member{CampaignID: camp, UserID: owner, Role: campaigns.RolePlayer}
	if err := r.svc.Use(ctx, ana, charA, PlaceBackpack, 0); err != nil {
		t.Fatal(err)
	}
	if len(r.chars.slot("backpack")) != 0 {
		t.Fatal("item not consumed")
	}
	if len(r.ef.got) != 1 || r.ef.got[0].Effect != effects.EffectHeal || r.ef.got[0].Amount != 2 || r.ef.got[0].Message != "Usou Bandagem: curar 2" {
		t.Fatalf("suggestions = %+v", r.ef.got)
	}
}

func TestUseRulesAndRefugeResource(t *testing.T) {
	r := newRig()
	tool, _ := r.svc.Create(ctx, camp, validItem())
	water, _ := r.svc.Create(ctx, camp, Item{Name: "Garrafa", Icon: "droplet", Color: "#0000ff", Kind: KindConsumable,
		Effects: ItemEffects{RefugeResource: &RefugeResource{Resource: "water", Amount: 3}}})
	_ = giveTo(r, tool.ID, PlaceBackpack, 1)
	_ = giveTo(r, water.ID, PlaceBackpack, 1)
	bob := campaigns.Member{CampaignID: camp, UserID: "u-bob", Role: campaigns.RolePlayer}
	if err := r.svc.Use(ctx, bob, charA, PlaceBackpack, 1); err == nil {
		t.Fatal("non-owner used an item")
	}
	ana := campaigns.Member{CampaignID: camp, UserID: owner, Role: campaigns.RolePlayer}
	if err := r.svc.Use(ctx, ana, charA, PlaceBackpack, 0); err == nil || len(r.chars.slot("backpack")) != 2 {
		t.Fatalf("tool must not be usable: %v", err)
	}
	if err := r.svc.Use(ctx, ana, charA, PlaceBackpack, 9); err == nil {
		t.Fatal("bad index accepted")
	}
	if err := r.svc.Use(ctx, ana, charA, PlaceBackpack, 1); err != nil || r.rf.water != 3 {
		t.Fatalf("refuge resource: %v %d", err, r.rf.water)
	}
}

func place(r *rig, itemID string, difficulty int, obj *string) Loot {
	l, err := r.svc.PlaceLoot(ctx, camp, PlaceInput{SceneID: sceneA, X: 3, Z: 4, ObjectID: obj, ItemID: itemID, Qty: 1, Difficulty: difficulty})
	if err != nil {
		panic(err)
	}
	return l
}

func TestHiddenLootNeverReachesPlayers(t *testing.T) {
	r := newRig()
	it, _ := r.svc.Create(ctx, camp, validItem())
	crate := "crate1"
	l := place(r, it.ID, 2, &crate)
	gm := campaigns.Member{CampaignID: camp, Role: campaigns.RoleGM}
	ana := campaigns.Member{CampaignID: camp, UserID: owner, Role: campaigns.RolePlayer}
	if list, _ := r.svc.ListLoot(ctx, gm, sceneA); len(list) != 1 {
		t.Fatalf("gm list = %d", len(list))
	}
	if list, _ := r.svc.ListLoot(ctx, ana, sceneA); len(list) != 0 {
		t.Fatalf("player list = %d", len(list))
	}
	if err := r.svc.RemoveLoot(ctx, camp, l.ID); err != nil {
		t.Fatal(err)
	}
	if seen := r.ev.playerSees(); len(seen) != 0 {
		t.Fatalf("player saw %v", seen)
	}
}

func TestPlaceLootValidation(t *testing.T) {
	r := newRig()
	it, _ := r.svc.Create(ctx, camp, validItem())
	bad := []PlaceInput{
		{SceneID: sceneA, X: 99, Z: 0, ItemID: it.ID},
		{SceneID: "other", X: 0, Z: 0, ItemID: it.ID},
		{SceneID: sceneA, X: 0, Z: 0, ItemID: it.ID, Difficulty: -1},
		{SceneID: sceneA, X: 0, Z: 0, ItemID: "nope"},
		{SceneID: sceneA, X: 0, Z: 0, ItemID: it.ID, ObjectID: strPtr("ghost")},
	}
	for i, in := range bad {
		if _, err := r.svc.PlaceLoot(ctx, camp, in); err == nil {
			t.Errorf("case %d accepted", i)
		}
	}
}

func strPtr(s string) *string { return &s }

func TestTryFindByDifficultyAndFullBackpack(t *testing.T) {
	r := newRig()
	it, _ := r.svc.Create(ctx, camp, validItem())
	crate := "crate1"
	easy, hard := place(r, it.ID, 2, &crate), place(r, it.ID, 3, &crate)
	target := Target{ObjectID: "crate1"}
	found, err := r.svc.TryFind(ctx, camp, charA, sceneA, target, 2)
	if err != nil || len(found) != 1 || found[0].Loot.ID != easy.ID || found[0].Delivered != 1 {
		t.Fatalf("found = %+v %v", found, err)
	}
	if len(r.chars.slot("backpack")) != 1 {
		t.Fatal("item not in backpack")
	}
	if _, ok := r.repo.loot[hard.ID]; !ok || !r.repo.loot[hard.ID].Hidden {
		t.Fatal("harder loot must stay hidden")
	}
	if got := eventTypes(r.ev, EventLootFound); got != 1 {
		t.Fatalf("loot.found events = %d", got)
	}
	for i := 0; i < 5; i++ { // fill the backpack
		_ = giveTo(r, it.ID, PlaceBackpack, 1)
	}
	found, _ = r.svc.TryFind(ctx, camp, charA, sceneA, Target{Cells: []maps.Coord{{X: 3, Z: 4}}}, 5)
	if len(found) != 1 || found[0].Delivered != 0 || found[0].Left != 1 {
		t.Fatalf("full backpack found = %+v", found)
	}
	if l := r.repo.loot[hard.ID]; l.Hidden || l.FoundBy == nil {
		t.Fatalf("leftover must lie visible: %+v", l)
	}
}

func eventTypes(ev *fakeEvents, typ string) int {
	n := 0
	for _, e := range ev.list {
		if e.typ == typ {
			n++
		}
	}
	return n
}

func TestPickupNeedsAdjacencyAndVisibility(t *testing.T) {
	r := newRig()
	it, _ := r.svc.Create(ctx, camp, validItem())
	ana := campaigns.Member{CampaignID: camp, UserID: owner, Role: campaigns.RolePlayer}
	hidden := place(r, it.ID, 1, nil)
	if err := r.svc.Pickup(ctx, ana, hidden.ID, charA); err == nil {
		t.Fatal("hidden loot picked up")
	}
	visible, _ := r.svc.PlaceLoot(ctx, camp, PlaceInput{SceneID: sceneA, X: 3, Z: 4, ItemID: it.ID, Qty: 1, Hidden: boolPtr(false)})
	r.tk.x, r.tk.z = 6, 4
	if err := r.svc.Pickup(ctx, ana, visible.ID, charA); err == nil || !strings.Contains(err.Error(), "perto") {
		t.Fatalf("far pickup err = %v", err)
	}
	r.tk.x = 4
	if err := r.svc.Pickup(ctx, ana, visible.ID, charA); err != nil {
		t.Fatal(err)
	}
	if len(r.chars.slot("backpack")) != 1 || len(r.repo.loot) != 1 {
		t.Fatalf("pickup state: %d loot=%d", len(r.chars.slot("backpack")), len(r.repo.loot))
	}
}

func boolPtr(b bool) *bool { return &b }
