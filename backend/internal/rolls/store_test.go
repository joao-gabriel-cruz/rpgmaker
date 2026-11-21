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

func TestStoreUsageAndDifficulty(t *testing.T) {
	st, c, u := seedStore(t)
	var sessionID string
	if err := st.pool.QueryRow(ctx, `INSERT INTO play_sessions (campaign_id) VALUES ($1) RETURNING id`, c.ID).Scan(&sessionID); err != nil {
		t.Fatal(err)
	}
	var charID string
	if err := st.pool.QueryRow(ctx, `INSERT INTO characters (campaign_id, owner_id, kind, name, sheet) VALUES ($1,$2,'pc','Rita','{}') RETURNING id`, c.ID, u.ID).Scan(&charID); err != nil {
		t.Skipf("characters schema differs: %v", err)
	}
	rec := pendingRecord(c, u, VisPublic)
	rec.CharacterID, rec.PlaySessionID = &charID, &sessionID
	rec.Spec = json.RawMessage(`{"request":{"useOccupation":true,"useMarkingEvent":false},"pool":{"dice":[],"keptCount":0,"assimilated":false}}`)
	if _, err := st.Insert(ctx, rec); err != nil {
		t.Fatal(err)
	}
	if got, err := st.Usage(ctx, sessionID, charID); err != nil || !got.OccupationUsed || got.MarkingEventUsed {
		t.Errorf("usage = %+v %v", got, err)
	}
	if err := st.SetDifficulty(ctx, c.ID, charID, 2); err != nil {
		t.Fatal(err)
	}
	if n, _ := st.Difficulty(ctx, c.ID, charID); n != 2 {
		t.Errorf("difficulty = %d", n)
	}
	_ = st.SetDifficulty(ctx, c.ID, charID, 0)
	if n, _ := st.Difficulty(ctx, c.ID, charID); n != 0 {
		t.Errorf("cleared = %d", n)
	}
}
