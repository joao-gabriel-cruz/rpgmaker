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
	if err != nil || r.Name != "Refúgio" || len(r.Items) != 0 || r.Population.Max != 10 || r.ReservesMax != 10 || r.Week != 1 || len(r.Resources) != 12 {
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
}

func TestStoreMutateKeepsEverythingAndRollsBack(t *testing.T) {
	st, _, campID, _ := seedDB(t)
	if _, err := st.Mutate(ctx, campID, func(r *Refuge) error {
		r.Name, r.Defense, r.Resources["wood"], r.Region.Abundance["water"] = "Base", 2, 4, "abundant"
		r.Items = append(r.Items, json.RawMessage(`{"name":"Faca","itemId":"x"}`))
		r.Crises = []json.RawMessage{json.RawMessage(`{"id":"c1"}`)}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("boom")
	if _, err := st.Mutate(ctx, campID, func(r *Refuge) error { r.Defense = 6; return boom }); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	r, _ := st.Get(ctx, campID)
	if r.Name != "Base" || r.Defense != 2 || r.Resources["wood"] != 4 || r.Region.Abundance["water"] != "abundant" ||
		len(r.Items) != 1 || !strings.Contains(string(r.Items[0]), "itemId") || len(r.Crises) != 1 || len(r.Buildings) != 0 {
		t.Errorf("state = %+v", r)
	}
}

const oldSchema = `DROP SCHEMA IF EXISTS mig_test CASCADE; CREATE SCHEMA mig_test; SET search_path TO mig_test;
	CREATE TABLE refuges (campaign_id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL DEFAULT 'Refúgio',
	description text NOT NULL DEFAULT '', defense int NOT NULL DEFAULT 0, morale int NOT NULL DEFAULT 0,
	mobility int NOT NULL DEFAULT 0, belligerence int NOT NULL DEFAULT 0, water int NOT NULL DEFAULT 0,
	food int NOT NULL DEFAULT 0, wood int NOT NULL DEFAULT 0, notes text NOT NULL DEFAULT '',
	items jsonb NOT NULL DEFAULT '[]', updated_at timestamptz NOT NULL DEFAULT now());
	INSERT INTO refuges (name, defense, morale, mobility, belligerence, water, food, wood, items)
	VALUES ('Velho', 10, 5, 7, 1, 8, 2, 30, '[{"name":"Faca"}]');`

// migrateOld applies the 00078 Up section to an old-format table in a scratch schema.
func migrateOld(t *testing.T) (name, state, items string) {
	t.Helper()
	c, err := testPool(t).Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = c.Exec(ctx, `SET search_path TO public; DROP SCHEMA IF EXISTS mig_test CASCADE`)
		c.Release()
	})
	raw, err := os.ReadFile("../migrations/sql/00078_refuge_v2.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.Split(strings.TrimPrefix(string(raw), "-- +goose Up"), "-- +goose Down")[0]
	if _, err := c.Exec(ctx, oldSchema+up); err != nil {
		t.Fatal(err)
	}
	if err := c.QueryRow(ctx, `SELECT name, state::text, items::text FROM refuges`).Scan(&name, &state, &items); err != nil {
		t.Fatal(err)
	}
	return name, state, items
}

func TestMigrationConvertsOldRefuge(t *testing.T) {
	name, state, items := migrateOld(t)
	var r Refuge
	if err := decodeState([]byte(state), &r); err != nil {
		t.Fatal(err)
	}
	if name != "Velho" || r.Defense != 6 || r.Morale != 3 || r.Belligerence != 1 || r.Mobility != 7 ||
		r.Resources["water"] != 8 || r.Resources["food"] != 2 || r.Resources["wood"] != 30 || r.ReservesMax != 40 ||
		r.Population.Max != 10 || r.Week != 1 || !strings.Contains(items, "Faca") {
		t.Errorf("migrated = %s %+v items=%s", name, r.Sheet, items)
	}
}
