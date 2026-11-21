package mcp

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
)

func TestLinkStoreReplaceKeepsOneActive(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	c := context.Background()
	pool, err := database.Connect(c, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrations.Up(c, database.StdDB(pool)); err != nil {
		t.Fatal(err)
	}
	_, _ = pool.Exec(c, `TRUNCATE users CASCADE`)
	u, _ := auth.NewStore(pool).CreateUser(c, "gm@x.com", "h", "GM", false)
	cp, _ := campaigns.NewStore(pool).CreateCampaign(c, "T", "assimilacao", u.ID)
	s := NewLinkRepo(pool)
	if _, err := s.Replace(c, cp.ID, u.ID, "h1"); err != nil {
		t.Fatal(err)
	}
	second, err := s.Replace(c, cp.ID, u.ID, "h2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ByHash(c, "h1"); !errors.Is(err, ErrInvalidLink) {
		t.Fatalf("h1 should be revoked: %v", err)
	}
	if a, err := s.Active(c, cp.ID, u.ID); err != nil || a.ID != second.ID {
		t.Fatalf("active: %+v %v", a, err)
	}
	if err := s.Touch(c, second.ID); err != nil {
		t.Fatal(err)
	}
	_ = s.Revoke(c, cp.ID, u.ID)
	if _, err := s.Active(c, cp.ID, u.ID); !errors.Is(err, ErrNoLink) {
		t.Fatalf("after revoke: %v", err)
	}
}
