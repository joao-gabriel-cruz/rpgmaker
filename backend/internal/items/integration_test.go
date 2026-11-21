package items

import (
	"encoding/json"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/refuge"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

// TestTryFindEndToEndOnPostgres is what the Perceber/Vasculhar action will do: 2 Successes find a difficulty-2
// loot hidden in a crate, announce loot.found to everyone and put the item in the finder's backpack.
func TestTryFindEndToEndOnPostgres(t *testing.T) {
	pool := testPool(t)
	authSvc := auth.NewService(auth.NewStore(pool))
	u, _ := authSvc.CreateUser(ctx, "kaio-items@x.com", "Kaio", "password123", false)
	camp, _ := campaigns.NewStore(pool).CreateCampaign(ctx, "T", "assimilacao", u.ID)
	ev := &fakeEvents{}
	chars := characters.NewService(characters.NewStore(pool), campaigns.NewService(campaigns.NewStore(pool), authSvc, systems.Default()), systems.Default(), ev)
	kaio, err := chars.Create(ctx, campaigns.Member{CampaignID: camp.ID, UserID: u.ID, Role: campaigns.RoleGM}, characters.CreateInput{Name: "Kaio", OwnerID: u.ID})
	if err != nil {
		t.Fatal(err)
	}
	sceneSvc := scenes.NewService(scenes.NewStore(pool), ev)
	sc, _ := sceneSvc.Create(ctx, camp.ID, "S", 16, 16)
	if err := sceneSvc.EditObjects(ctx, camp.ID, sc.ID, "", scenes.ObjectPatch{Add: []maps.Object{{ID: "crate-1", Prefab: "crate", X: 5, Z: 5}}}); err != nil {
		t.Fatal(err)
	}
	svc := NewService(NewStore(pool), chars, refuge.NewService(refuge.NewStore(pool), chars, ev), &fakeEffects{}, &fakeTokens{}, sceneSvc, ev)
	it, _ := svc.Create(ctx, camp.ID, validItem())
	crate := "crate-1"
	if _, err := svc.PlaceLoot(ctx, camp.ID, PlaceInput{SceneID: sc.ID, X: 5, Z: 5, ObjectID: &crate, ItemID: it.ID, Qty: 1, Difficulty: 2}); err != nil {
		t.Fatal(err)
	}
	if f, _ := svc.TryFind(ctx, camp.ID, kaio.ID, sc.ID, Target{ObjectID: crate}, 1); len(f) != 0 {
		t.Fatalf("1 Success must not find a difficulty-2 loot: %+v", f)
	}
	found, err := svc.TryFind(ctx, camp.ID, kaio.ID, sc.ID, Target{ObjectID: crate}, 2)
	if err != nil || len(found) != 1 || found[0].Delivered != 1 {
		t.Fatalf("found = %+v %v", found, err)
	}
	if eventTypes(ev, EventLootFound) != 1 || !containsType(ev.playerSees(), EventLootFound) {
		t.Fatalf("loot.found must reach everyone: %v", ev.playerSees())
	}
	view, _ := chars.Get(ctx, campaigns.Member{CampaignID: camp.ID, UserID: u.ID, Role: campaigns.RoleGM}, kaio.ID)
	var sheet struct {
		Inventory struct {
			Backpack []struct{ Name, ItemID string }
		}
	}
	_ = json.Unmarshal(view.Sheet, &sheet)
	if len(sheet.Inventory.Backpack) != 1 || sheet.Inventory.Backpack[0].ItemID != it.ID {
		t.Fatalf("backpack = %s", view.Sheet)
	}
	if list, _ := svc.ListLoot(ctx, campaigns.Member{CampaignID: camp.ID, Role: campaigns.RoleGM}, sc.ID); len(list) != 0 {
		t.Fatal("delivered loot must be gone")
	}
}

func containsType(list []string, typ string) bool {
	for _, s := range list {
		if s == typ {
			return true
		}
	}
	return false
}
