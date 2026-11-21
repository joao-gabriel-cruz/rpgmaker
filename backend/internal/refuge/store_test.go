package refuge

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
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
	if _, err := pool.Exec(ctx, `TRUNCATE users CASCADE`); err != nil {
		t.Fatal(err)
	}
	return pool
}

func seedDB(t *testing.T) (*Store, *pgxpool.Pool, string, string) {
	t.Helper()
	pool := testPool(t)
	authSvc := auth.NewService(auth.NewStore(pool))
	u, err := authSvc.CreateUser(ctx, "g@x.com", "g", "g-password-1", false)
	if err != nil {
		t.Fatal(err)
	}
	camp, _ := campaigns.NewStore(pool).CreateCampaign(ctx, "T", "assimilacao", u.ID)
	var cid string
	if err := pool.QueryRow(ctx, `INSERT INTO characters (campaign_id, owner_id, kind, name, sheet)
		VALUES ($1,$2,'pc','R',$3) RETURNING id`, camp.ID, u.ID,
		`{"inventory":{"body":[],"backpack":[{"name":"Faca"}]}}`).Scan(&cid); err != nil {
		t.Fatal(err)
	}
	return NewStore(pool), pool, camp.ID, cid
}

func TestStoreMoveIsTransactional(t *testing.T) {
	st, pool, campID, cid := seedDB(t)
	r, err := st.Get(ctx, campID)
	if err != nil || r.Name != "Refúgio" || len(r.Items) != 0 {
		t.Fatalf("get = %+v %v", r, err)
	}
	if _, err := st.Move(ctx, campID, cid, func(c CharState, i []json.RawMessage) (json.RawMessage, []json.RawMessage, error) {
		return deposit(c, i, "backpack", 0)
	}); err != nil {
		t.Fatal(err)
	}
	failing := func(c CharState, i []json.RawMessage) (json.RawMessage, []json.RawMessage, error) {
		return withdraw(c, i, "body", 5)
	}
	if _, err := st.Move(ctx, campID, cid, failing); err == nil {
		t.Fatal("expected error")
	}
	r, _ = st.Get(ctx, campID)
	var sheet string
	_ = pool.QueryRow(ctx, `SELECT sheet::text FROM characters WHERE id=$1`, cid).Scan(&sheet)
	if len(r.Items) != 1 || strings.Contains(sheet, "Faca") {
		t.Errorf("items=%v sheet=%s", r.Items, sheet)
	}
	if _, err := st.Move(ctx, campID, "00000000-0000-0000-0000-000000000000", failing); !errors.Is(err, ErrCharNotFound) {
		t.Errorf("missing char = %v", err)
	}
	if r, err := st.Update(ctx, campID, Details{Name: "Base", Water: 4, Defense: 2}); err != nil || r.Water != 4 || len(r.Items) != 1 {
		t.Errorf("update = %+v %v", r, err)
	}
}
