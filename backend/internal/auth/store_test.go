package auth

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
)

func testStore(t *testing.T) *Store {
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
	return NewStore(pool)
}

func TestStoreUsers(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	u, err := s.CreateUser(ctx, "Ana@Example.com", "hash", "Ana", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateUser(ctx, "ana@example.COM", "h", "Dup", false); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("duplicate: %v", err)
	}
	if got, err := s.UserByEmail(ctx, "ANA@example.com"); err != nil || got.ID != u.ID {
		t.Errorf("by email: %v %v", got, err)
	}
	if _, err := s.UserByEmail(ctx, "no@x.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing: %v", err)
	}
	users, err := s.ListUsers(ctx)
	if err != nil || len(users) != 1 {
		t.Errorf("list: %v %v", users, err)
	}
}

func TestStoreSessions(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	u, err := s.CreateUser(ctx, "ana@example.com", "hash", "Ana", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, "live", u.ID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, "dead", u.ID, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got, err := s.UserBySession(ctx, "live"); err != nil || got.ID != u.ID {
		t.Errorf("live session: %v %v", got, err)
	}
	if _, err := s.UserBySession(ctx, "dead"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expired session: %v", err)
	}
	if err := s.SetPassword(ctx, u.ID, "newhash"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UserBySession(ctx, "live"); !errors.Is(err, ErrNotFound) {
		t.Errorf("sessions should be revoked: %v", err)
	}
	if err := s.SetPassword(ctx, "00000000-0000-0000-0000-000000000000", "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("set password missing: %v", err)
	}
}
