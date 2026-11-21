package campaigns

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
)

func testStores(t *testing.T) (*Store, *auth.Store) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
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
	return NewStore(pool), auth.NewStore(pool)
}

func TestStoreCampaignMembersAndInvites(t *testing.T) {
	s, users := testStores(t)
	ctx := context.Background()
	gm, _ := users.CreateUser(ctx, "gm@x.com", "h", "GM", false)
	pl, _ := users.CreateUser(ctx, "pl@x.com", "h", "PL", false)

	c, err := s.CreateCampaign(ctx, "Teste", "assimilacao", gm.ID)
	if err != nil || string(c.Settings) != "{}" {
		t.Fatalf("create: %+v %v", c, err)
	}
	if m, err := s.Member(ctx, c.ID, gm.ID); err != nil || m.Role != RoleGM {
		t.Errorf("owner member: %+v %v", m, err)
	}
	if _, err := s.Member(ctx, c.ID, pl.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("non-member: %v", err)
	}

	seedInvites(t, s, c.ID)
	if _, err := s.ConsumeInvite(ctx, "expired", pl.ID, time.Now()); !errors.Is(err, ErrInvalidInvite) {
		t.Errorf("expired: %v", err)
	}
	if m, err := s.ConsumeInvite(ctx, "good", pl.ID, time.Now()); err != nil || m.Role != RolePlayer {
		t.Errorf("consume: %+v %v", m, err)
	}
	if _, err := s.ConsumeInvite(ctx, "good", pl.ID, time.Now()); !errors.Is(err, ErrInvalidInvite) {
		t.Errorf("reuse: %v", err)
	}
	checkMembership(t, s, c.ID, pl.ID)
}

func seedInvites(t *testing.T, s *Store, campaignID string) {
	t.Helper()
	exp := map[string]time.Time{"good": time.Now().Add(time.Hour), "expired": time.Now().Add(-time.Hour)}
	for tok, at := range exp {
		inv := Invite{Token: tok, CampaignID: campaignID, Role: RolePlayer, ExpiresAt: at}
		if err := s.CreateInvite(context.Background(), inv); err != nil {
			t.Fatal(err)
		}
	}
}

func checkMembership(t *testing.T, s *Store, campaignID, playerID string) {
	t.Helper()
	ctx := context.Background()
	if list, err := s.ListByUser(ctx, playerID); err != nil || len(list) != 1 || list[0].Role != RolePlayer {
		t.Errorf("list: %+v %v", list, err)
	}
	if ms, err := s.Members(ctx, campaignID); err != nil || len(ms) != 2 || ms[0].Role != RoleGM {
		t.Errorf("members: %+v %v", ms, err)
	}
	if err := s.RemoveMember(ctx, campaignID, playerID); err != nil {
		t.Errorf("remove: %v", err)
	}
	if err := s.RemoveMember(ctx, campaignID, playerID); !errors.Is(err, ErrNotFound) {
		t.Errorf("remove twice: %v", err)
	}
}
