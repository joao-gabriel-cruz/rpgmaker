package drawings

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

func seed(t *testing.T) (pool *pgxpool.Pool, campID, sceneID, userID string) {
	t.Helper()
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
	cp, err := campaigns.NewStore(pool).CreateCampaign(c, "T", "assimilacao", u.ID)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := scenes.NewStore(pool).Create(c, cp.ID, "S", 16, 16, strings.Repeat("0", 16*16*2))
	if err != nil {
		t.Fatal(err)
	}
	return pool, cp.ID, sc.ID, u.ID
}

func TestStoreRoundTrip(t *testing.T) {
	pool, campID, sceneID, userID := seed(t)
	c := context.Background()
	s := NewStore(pool)
	d, err := s.Add(c, Drawing{SceneID: sceneID, UserID: userID, Color: "#00ff00", Width: 2, Points: stroke(3)})
	if err != nil || d.ID == "" || len(d.Points) != 3 {
		t.Fatalf("add: %+v %v", d, err)
	}
	got, err := s.Get(c, campID, d.ID)
	list, _ := s.List(c, campID, sceneID)
	if err != nil || got.Color != "#00ff00" || len(list) != 1 {
		t.Fatalf("get/list: %+v %v %v", got, err, list)
	}
	if err := s.Remove(c, d.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(c, d.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	_, _ = s.Add(c, Drawing{SceneID: sceneID, UserID: userID, Color: "#00ff00", Width: 2, Points: stroke(2)})
	if err := s.Clear(c, sceneID); err != nil {
		t.Fatal(err)
	}
	if l, _ := s.List(c, campID, sceneID); len(l) != 0 {
		t.Fatalf("not cleared: %v", l)
	}
}
