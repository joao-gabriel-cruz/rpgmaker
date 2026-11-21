package items

import (
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
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

func seedCampaigns(t *testing.T, pool *pgxpool.Pool) (campID, otherID, sceneID string) {
	t.Helper()
	u, _ := auth.NewStore(pool).CreateUser(ctx, "gm-items@x.com", "h", "GM", false)
	cp, err := campaigns.NewStore(pool).CreateCampaign(ctx, "T", "assimilacao", u.ID)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := campaigns.NewStore(pool).CreateCampaign(ctx, "O", "assimilacao", u.ID)
	g, _ := maps.NewGrid(16, 16)
	sc, err := scenes.NewStore(pool).Create(ctx, cp.ID, "S", 16, 16, g.Serialize().Data)
	if err != nil {
		t.Fatal(err)
	}
	return cp.ID, other.ID, sc.ID
}

func TestStoreItemsAndLoot(t *testing.T) {
	pool := testPool(t)
	campID, otherID, sceneID := seedCampaigns(t, pool)
	st := NewStore(pool)
	it, err := st.CreateItem(ctx, campID, validItem())
	if err != nil || it.ID == "" || it.Effects.KeptBonus == nil {
		t.Fatalf("create: %+v %v", it, err)
	}
	if _, err := st.GetItem(ctx, otherID, it.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-campaign get = %v", err)
	}
	it.Name = "Barra"
	if up, err := st.UpdateItem(ctx, campID, it); err != nil || up.Name != "Barra" {
		t.Fatalf("update: %v", err)
	}
	checkLootLifecycle(t, st, campID, otherID, sceneID, it)
}

func checkLootLifecycle(t *testing.T, st *Store, campID, otherID, sceneID string, it Item) {
	t.Helper()
	l, err := st.InsertLoot(ctx, Loot{SceneID: sceneID, X: 1, Z: 2, ItemID: &it.ID, Item: SheetItem(it), Qty: 2, Difficulty: 2, Hidden: true})
	if err != nil || l.ID == "" || !l.Hidden {
		t.Fatalf("insert loot: %+v %v", l, err)
	}
	if list, _ := st.ListLoot(ctx, otherID, sceneID); len(list) != 0 {
		t.Fatal("loot leaked across campaigns")
	}
	l.Qty, l.Hidden = 1, false
	if err := st.SaveLoot(ctx, l); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.GetLoot(ctx, campID, l.ID); got.Qty != 1 || got.Hidden {
		t.Fatalf("save: %+v", got)
	}
	if err := st.DeleteItem(ctx, campID, it.ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.GetLoot(ctx, campID, l.ID); got.ItemID != nil || len(got.Item) == 0 {
		t.Fatalf("loot must survive item deletion with snapshot: %+v", got)
	}
	if err := st.DeleteLoot(ctx, campID, l.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteLoot(ctx, campID, l.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete = %v", err)
	}
}
