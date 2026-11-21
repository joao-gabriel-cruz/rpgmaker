package rolls

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
)

func seedStore(t *testing.T) (*Store, campaigns.Campaign, auth.User) {
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
	authSvc := auth.NewService(auth.NewStore(pool))
	u, err := authSvc.CreateUser(ctx, "gm@x.com", "gm", "gm-password", false)
	if err != nil {
		t.Fatal(err)
	}
	campStore := campaigns.NewStore(pool)
	c, _ := campStore.CreateCampaign(ctx, "T", "assimilacao", u.ID)
	return NewStore(pool), c, u
}

func pendingRecord(c campaigns.Campaign, u auth.User, vis string) Record {
	return Record{CampaignID: c.ID, UserID: u.ID, Visibility: vis, Kind: KindAssimilacao, Status: StatusPending,
		Spec: json.RawMessage(`{"request":{},"pool":{"dice":[],"keptCount":0,"assimilated":false}}`), Dice: json.RawMessage(`[]`), Seed: 7}
}

func TestStoreVisibilityFilter(t *testing.T) {
	st, c, u := seedStore(t)
	mk := func(vis string) Record { return pendingRecord(c, u, vis) }
	pub, err := st.Insert(ctx, mk(VisPublic))
	if err != nil || pub.UserName != "gm" || pub.Status != StatusPending || pub.Seed != 7 {
		t.Fatalf("insert = %+v %v", pub, err)
	}
	if _, err := st.Insert(ctx, mk(VisGM)); err != nil {
		t.Fatal(err)
	}
	if l, _ := st.Recent(ctx, c.ID, 50, false); len(l) != 1 {
		t.Errorf("player list = %d", len(l))
	}
	if l, _ := st.Recent(ctx, c.ID, 50, true); len(l) != 2 {
		t.Errorf("gm list = %d", len(l))
	}
}

func TestStoreFinalize(t *testing.T) {
	st, c, u := seedStore(t)
	pub, err := st.Insert(ctx, pendingRecord(c, u, VisPublic))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Finalize(ctx, c.ID, pub.ID, json.RawMessage(`{"penalty":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := st.Finalize(ctx, c.ID, pub.ID, json.RawMessage(`{}`)); !errors.Is(err, ErrAlreadyFinal) {
		t.Errorf("double finalize = %v", err)
	}
	if got, _ := st.Get(ctx, c.ID, pub.ID); got.Status != StatusFinal || string(got.Result) != `{"penalty": 1}` {
		t.Errorf("got = %+v", got)
	}
	if _, err := st.Get(ctx, c.ID, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad id = %v", err)
	}
}
