package conflict

import (
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
)

func seed(t *testing.T) (*pgxpool.Pool, campaigns.Campaign, auth.User) {
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
	u, err := auth.NewService(auth.NewStore(pool)).CreateUser(ctx, "gm@x.com", "gm", "gm-password", false)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := campaigns.NewStore(pool).CreateCampaign(ctx, "T", "assimilacao", u.ID)
	return pool, c, u
}

func TestStoreLifecycle(t *testing.T) {
	pool, c, u := seed(t)
	st := NewStore(pool)
	saved, err := st.Save(ctx, Conflict{CampaignID: c.ID, Title: "A", Doc: Doc{Conditions: []string{"x"}}})
	if err != nil || saved.ID == "" {
		t.Fatalf("save: %v", err)
	}
	saved.Title = "B"
	if _, err := st.Save(ctx, saved); err != nil {
		t.Fatal(err)
	}
	got, _ := st.Active(ctx, c.ID)
	if got == nil || got.Title != "B" || got.Conditions[0] != "x" {
		t.Fatalf("active = %+v", got)
	}
	rollID := threatRoll(t, pool, c.ID, u.ID)
	if err := st.Assign(ctx, saved.ID, rollID, "o1", 1); err != nil {
		t.Fatal(err)
	}
	if err := st.Assign(ctx, saved.ID, rollID, "o1", 1); !errors.Is(err, ErrAlreadyAssigned) {
		t.Fatalf("second assign: %v", err)
	}
	_ = st.Close(ctx, c.ID)
	if got, _ := st.Active(ctx, c.ID); got != nil {
		t.Fatal("closed conflict still active")
	}
}

func threatRoll(t *testing.T, pool *pgxpool.Pool, campaignID, userID string) string {
	t.Helper()
	rollSvc := rolls.NewService(rolls.NewStore(pool), nil, nil, noEvents{})
	if _, err := rollSvc.Threat(ctx, campaignID, senderOf(userID), "Horda", rolls.ThreatPool{D6: 2, D12: 1}); err != nil {
		t.Fatal(err)
	}
	recs, _ := rolls.NewStore(pool).Recent(ctx, campaignID, 5, true)
	if len(recs) != 1 || recs[0].Kind != rolls.KindThreat {
		t.Fatalf("threat roll not logged: %+v", recs)
	}
	return recs[0].ID
}
