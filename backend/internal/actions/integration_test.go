package actions

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/items"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/refuge"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/sessions"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := database.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrations.Up(ctx, database.StdDB(pool)); err != nil {
		t.Fatal(err)
	}
	_, _ = pool.Exec(ctx, `TRUNCATE users CASCADE`)
	return pool
}

func TestStoreRoundTripOnPostgres(t *testing.T) {
	pool := testPool(t)
	u, _ := auth.NewStore(pool).CreateUser(ctx, "gm-actions@x.com", "h", "GM", false)
	cp, err := campaigns.NewStore(pool).CreateCampaign(ctx, "T", "assimilacao", u.ID)
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(NewStore(pool), nil, nil, nil, nil, nil, nil, nil)
	d, _ := defaultByID("rest")
	d.Name = "Acampar"
	if _, err := svc.Update(ctx, cp.ID, "rest", d); err != nil {
		t.Fatal(err)
	}
	list, _ := svc.Catalog(ctx, cp.ID, true)
	if len(list) != 11 || list[7].Name != "Acampar" || !list[7].Builtin {
		t.Fatalf("catalog = %+v", list[7])
	}
	if err := svc.Delete(ctx, cp.ID, "rest"); err != nil {
		t.Fatal(err)
	}
	if a, _ := svc.Find(ctx, cp.ID, "rest"); a.Name != "Descansar" {
		t.Fatalf("restored = %+v", a)
	}
}

type pgRig struct {
	svc    *Service
	rolls  *rolls.Service
	ev     *fakeEvents
	camp   string
	char   string
	sender realtime.Sender
}

// newPgRig wires real characters, scenes, tokens, rolls, effects and items on Postgres with one placed PC and hidden loot nearby.
func newPgRig(t *testing.T) *pgRig {
	t.Helper()
	pool := testPool(t)
	au := auth.NewService(auth.NewStore(pool))
	u, _ := au.CreateUser(ctx, "ana-actions@x.com", "Ana", "password123", false)
	cp, _ := campaigns.NewStore(pool).CreateCampaign(ctx, "T", "assimilacao", u.ID)
	ev := &fakeEvents{}
	chars := characters.NewService(characters.NewStore(pool), campaigns.NewService(campaigns.NewStore(pool), au, systems.Default()), systems.Default(), ev)
	ana, err := chars.Create(ctx, campaigns.Member{CampaignID: cp.ID, UserID: u.ID, Role: campaigns.RoleGM}, characters.CreateInput{Name: "Ana", OwnerID: u.ID})
	if err != nil {
		t.Fatal(err)
	}
	sceneSvc := scenes.NewService(scenes.NewStore(pool), ev)
	sc, _ := sceneSvc.Create(ctx, cp.ID, "S", 16, 16)
	_ = sceneSvc.Activate(ctx, cp.ID, sc.ID)
	sender := realtime.Sender{UserID: u.ID, Role: campaigns.RoleGM}
	tokSvc := tokens.NewService(tokens.NewStore(pool), sceneSvc, ev)
	if err := tokSvc.Place(ctx, sender, cp.ID, sc.ID, ana.ID, 5, 5); err != nil {
		t.Fatal(err)
	}
	rollSvc := rolls.NewService(rolls.NewStore(pool), chars, sessions.NewService(sessions.NewStore(pool), ev), ev)
	fx := effects.NewService(effects.NewStore(pool), sceneSvc, tokens.NewStore(pool), chars, rollSvc, ev)
	itemSvc := items.NewService(items.NewStore(pool), chars, refuge.NewService(refuge.NewStore(pool), chars, ev), fx, tokSvc, sceneSvc, ev)
	hideLoot(t, itemSvc, cp.ID, sc.ID)
	svc := NewService(NewStore(pool), sceneSvc, tokens.NewStore(pool), chars, rollSvc, itemSvc, fx, ev)
	rollSvc.OnFinalized(svc.Finalized)
	return &pgRig{svc: svc, rolls: rollSvc, ev: ev, camp: cp.ID, char: ana.ID, sender: sender}
}

func hideLoot(t *testing.T, svc *items.Service, campID, sceneID string) {
	t.Helper()
	it, err := svc.Create(ctx, campID, items.Item{Name: "Lata", Icon: "package", Color: "#aaaaaa", Kind: items.KindOther})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PlaceLoot(ctx, campID, items.PlaceInput{SceneID: sceneID, X: 6, Z: 6, ItemID: it.ID, Qty: 1}); err != nil {
		t.Fatal(err)
	}
}

// TestPerformThroughRealRollsFlow runs Perceber over the real rolls and items services:
// roll.rolled -> roll.keep -> action.resolved, with the hidden loot revealed.
func TestPerformThroughRealRollsFlow(t *testing.T) {
	r := newPgRig(t)
	if err := r.svc.Perform(ctx, r.camp, r.sender, PerformInput{CharacterID: r.char, ActionID: "perceive"}); err != nil {
		t.Fatal(err)
	}
	started, _ := r.ev.last(EventStarted)
	rolled, _ := r.ev.last(rolls.EventRolled)
	entry := rolled.payload.(rolls.Entry)
	if started.payload.(map[string]any)["rollId"] != entry.RollID {
		t.Fatal("action.started must carry the roll id")
	}
	if err := r.rolls.Keep(ctx, r.camp, r.sender, entry.RollID, entry.Suggested); err != nil {
		t.Fatal(err)
	}
	res, ok := r.ev.last(EventResolved)
	if !ok || res.payload.(map[string]any)["outcome"].(outcome).Kind != KindLoot {
		t.Fatalf("resolved = %+v", res)
	}
	if _, ok := r.ev.last(items.EventLootFound); !ok {
		t.Fatal("hidden loot one cell away must be found by Perceber")
	}
}
